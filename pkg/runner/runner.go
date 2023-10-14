package runner

import (
	"Yi/pkg/db"
	"Yi/pkg/logging"
	"Yi/pkg/utils"
	"fmt"
	"os"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/thoas/go-funk"
)

/**
  @author: yhy
  @since: 2022/10/13
  @desc: //TODO
**/

func Run() {
	logging.Logger.Infoln("Yi Starting ... ")

	projects := make(chan db.Project, 100)

	go func() {
		if Option.Target != "" {
			Option.Target = strings.TrimRight(Option.Target, "/")
			name := utils.GetName(Option.Target)
			err, dbPath, res := GetRepos(Option.Target)

			project := db.Project{
				Project:       name,
				Url:           Option.Target,
				Language:      res.Language,
				DBPath:        dbPath,
				PushedAt:      res.PushedAt,
				Count:         0,
				DefaultBranch: res.DefaultBranch,
			}

			if err != nil {
				db.AddProject(project)
				return
			}

			projects <- project

		} else if Option.Targets != "" {
			targets := utils.LoadFile(Option.Targets)
			limit := make(chan bool, Option.Thread)
			var wg sync.WaitGroup

			for _, target := range targets {
				target = strings.TrimRight(target, "/")
				limit <- true
				wg.Add(1)
				go func(target string) {
					defer func() {
						wg.Done()
						<-limit
					}()

					name := utils.GetName(target)
					err, dbPath, res := GetRepos(target)

					project := db.Project{
						Project:  name,
						Url:      target,
						Language: res.Language,
						Count:    0,
					}

					if err != nil {
						logging.Logger.Errorf("Add err(%s[%s]):%v", target, res.Language, err)
						db.AddProject(project)
						return
					}

					project.DBPath = dbPath
					project.PushedAt = res.PushedAt
					project.DefaultBranch = res.DefaultBranch
					projects <- project

				}(target)
			}
			wg.Wait()
			close(limit)
		}

		close(projects)
	}()

	limit := make(chan bool, Option.Thread)
	var wg sync.WaitGroup

	for project := range projects {
		if project.DBPath == "" {
			continue
		}
		wg.Add(1)
		limit <- true
		go WgExec(project, &wg, limit)
	}

	wg.Wait()
	close(limit)

	// After all running, you will start trying the wrong projects
	IsRetry = true
}

func WgExec(project db.Project, wg *sync.WaitGroup, limit chan bool) {
	exist, p := db.Exist(project.Url)

	if exist {
		project.Id = p.Id
		project.Count = p.Count
	} else {
		id, count := db.AddProject(project)
		project.Id = id
		project.Count = count
	}

	Exec(project, nil)

	<-limit
	wg.Done()
}

var LocationMaps = make(map[string]bool)

func Exec(project db.Project, qls []string) {
	if !funk.Contains(Languages, project.Language) || project.DBPath == "" {
		logging.Logger.Debugf("(%s)The current language does not support(%s)/The database is empty(%s)", project.Project, project.Language, project.DBPath)
		return
	}

	logging.Logger.Debugln("start exec project: ", project.Project)

	analyze := Analyze(project.DBPath, project.Project, project.Language, qls)

	for fileName, res := range analyze {
		results := jsoniter.Get([]byte(res), "runs", 0, "results")

		if results.Size() > 0 {
			for i := 0; i < results.Size(); i++ {
				location := "{"

				msg := "ruleId: " + results.Get(i).Get("ruleId").ToString() + "\t "
				if results.Get(i).Get("locations").Size() > 0 {
					for j := 0; j < results.Get(i).Get("locations").Size(); j++ {
						msg += "locations: " + results.Get(i).Get("locations").Get(j).Get("physicalLocation", "artifactLocation", "uri").ToString() + "\t startLine: " + results.Get(i).Get("locations").Get(j).Get("physicalLocation", "region", "startLine").ToString() + "\t | "

						line := results.Get(i).Get("locations").Get(j).Get("physicalLocation", "region", "startLine").ToString()
						location += fmt.Sprintf("\"%s#L%s\":\"%s\",", results.Get(i).Get("locations").Get(j).Get("physicalLocation", "artifactLocation", "uri").ToString(), line, line)
					}
				}

				if results.Get(i).Get("relatedLocations").Size() > 0 {
					for j := 0; j < results.Get(i).Get("relatedLocations").Size(); j++ {
						msg += "relatedLocations: " + results.Get(i).Get("relatedLocations").Get(j).Get("physicalLocation", "artifactLocation", "uri").ToString() + "\t startLine: " + results.Get(i).Get("relatedLocations").Get(j).Get("physicalLocation", "region", "startLine").ToString() + "\t | "

						line := results.Get(i).Get("relatedLocations").Get(j).Get("physicalLocation", "region", "startLine").ToString()
						location += fmt.Sprintf("\"%s#L%s\":\"%s\",", results.Get(i).Get("relatedLocations").Get(j).Get("physicalLocation", "artifactLocation", "uri").ToString(), line, line)
					}
				}

				location += "}"

				if _, ok := LocationMaps[location]; ok {
					continue
				}

				vul := db.Vul{
					Project:  project.Project,
					RuleId:   results.Get(i).Get("ruleId").ToString(),
					Location: []byte(location),
					//CodeFlows:     results.Get(i).Get("codeFlows").ToString(),
					Url:           project.Url,
					ResDir:        fileName,
					PushedAt:      project.PushedAt,
					DefaultBranch: project.DefaultBranch,
				}

				db.AddVul(vul)

				db.UpdateProjectArg(project.Id, "vul", 1)
				logging.Logger.Infof("%s(%s) Found: %s", project.Project, fileName, results.Get(i).Get("ruleId").ToString())
			}
		} else {
			err := os.Remove(fileName) //Delete Files

			if err != nil {
				logging.Logger.Infof("file remove Error! %s\n", err)
			}
		}
	}
	db.UpdateProjectArg(project.Id, "count", project.Count+1)
}

func ApiAdd(target, tag string) {
	var exist bool
	var project db.Project
	exist, project = db.Exist(target)
	if !exist {
		name := utils.GetName(target)
		err, dbPath, res := GetRepos(target)
		if err != nil {
			record := db.Record{
				Project: target,
				Url:     target,
				Color:   "danger",
				Title:   target,
				Msg:     fmt.Sprintf("%s add failed %v", target, err),
			}
			db.AddRecord(record)
			return
		}
		project = db.Project{
			Project:       name,
			DBPath:        dbPath,
			Url:           target,
			Tag:           tag,
			Language:      res.Language,
			PushedAt:      res.PushedAt,
			DefaultBranch: res.DefaultBranch,
			Count:         0,
		}
		id, _ := db.AddProject(project)
		project.Id = id
	} else {
		project.Count += 1

		record := db.Record{
			Project: project.Project,
			Url:     project.Url,
			Color:   "success",
			Title:   project.Project,
			Msg:     fmt.Sprintf("%s Already, reorganize", project.Url),
		}
		db.AddRecord(record)
	}

	Exec(project, nil)
}
