package db

import (
	"fmt"

	"gorm.io/gorm"
)

/**
  @author: yhy
  @since: 2022/12/7
  @desc: //TODO
**/

type Project struct {
	gorm.Model
	Id            int    `gorm:"primary_key" json:"id"`
	Project       string `json:"project"`
	Url           string `json:"url"`
	Language      string `json:"language"`
	Tag           string `json:"tag"`
	DBPath        string `json:"db_path"`
	Count         int    `json:"count"`
	PushedAt      string `json:"pushed_at"`
	Vul           int    `json:"vul"`
	DefaultBranch string `json:"default_branch"`
	LastScanTime  string
	ProgressBar   string `json:"progress_bar"`
}

func AddProject(project Project) (int, int) {
	GlobalDB.Create(&project)

	record := Record{
		Project: project.Project,
		Url:     project.Url,
		Color:   "success",
		Title:   project.Project,
		Msg:     fmt.Sprintf("%s Add successfully", project.Url),
	}
	AddRecord(record)

	return project.Id, project.Count
}

// GetProjects View Program Information
func GetProjects(pageNum int, pageSize int, maps interface{}) (count int64, projects []Project) {
	globalDBTmp := GlobalDB.Model(&Project{})
	query := maps.(map[string]interface{})

	if query["project"] != nil {
		globalDBTmp = globalDBTmp.Where("project LIKE ?", "%"+query["project"].(string)+"%")
	}

	if query["language"] != nil {
		globalDBTmp = globalDBTmp.Where("language LIKE ?", "%"+query["language"].(string)+"%")
	}

	globalDBTmp.Count(&count)
	if pageNum == 0 && pageSize == 0 {
		globalDBTmp.Find(&projects)

	} else {
		globalDBTmp.Offset(pageNum).Limit(pageSize).Order("vul desc,count desc").Find(&projects)
	}
	return
}

// UpdateProjectArg Update Fields
func UpdateProjectArg(id int, arg string, count int) bool {
	globalDBTmp := GlobalDB.Model(&Project{})
	globalDBTmp.Where("id = ?", id).Update(arg, count)
	return true
}

func DeleteProject(id string) {
	globalDBTmp := GlobalDB.Model(&Project{})
	globalDBTmp.Where("id = ?", id).Unscoped().Delete(&Project{})
}

// Exist  Determine whether ip and port exist in the database
func Exist(url string) (bool, Project) {
	var project Project
	globalDBTmp := GlobalDB.Model(&Project{})
	globalDBTmp.Where("url = ? ", url).Limit(1).First(&project)

	if project.Id > 0 {
		return true, project
	}

	return false, project
}

func UpdateProject(id int, project Project) {
	globalDBTmp := GlobalDB.Model(&Project{})
	globalDBTmp.Where("id = ?", id).Updates(project)
}
