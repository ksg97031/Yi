package db

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/**
  @author: yhy
  @since: 2022/10/18
  @desc: //TODO
**/

type Vul struct {
	gorm.Model
	Id            int    `gorm:"primary_key" json:"id"`
	Project       string `json:"project"`
	RuleId        string `json:"rule_id"`
	Url           string `json:"url"`
	DefaultBranch string `json:"default_branch"`
	PushedAt      string `json:"pushed_at"`
	Location      datatypes.JSON
	//CodeFlows     string `gorm:"type:text"`
	ResDir  string `json:"res_dir"`
	Handled bool   `json:"handled"`
}

func AddVul(vul Vul) {
	if ExistBlacklist(vul.Location.String()) {
		return
	}

	record := Record{
		Project: vul.Project,
		Url:     vul.Url,
		Color:   "danger",
		Title:   vul.Project + " Vulnerabilities found",
		Msg:     fmt.Sprintf("Vulnerability Type: %s", vul.RuleId),
	}
	AddRecord(record)

	t, _ := time.Parse(time.RFC3339, vul.PushedAt)
	vul.PushedAt = t.Format("2006-01-02 15:04:05")

	GlobalDB.Create(&vul)
}

// GetVulsHandled View Vulnerability Information
func GetVulsHandled(pageNum int, pageSize int, maps interface{}) (count int64, vuls []Vul) {
	globalDBTmp := GlobalDB.Model(&Vul{})
	query := maps.(map[string]interface{})

	if query["project"] != nil {
		globalDBTmp = globalDBTmp.Where("project LIKE ?", "%"+query["project"].(string)+"%")
	}

	if query["rule_id"] != nil {
		globalDBTmp = globalDBTmp.Where("rule_id LIKE ?", "%"+query["rule_id"].(string)+"%")
	}

	globalDBTmp.Where("handled = 1").Count(&count)

	globalDBTmp.Offset(pageNum).Limit(pageSize).Order("id asc").Find(&vuls)

	return
}

func GetVulsUnHandled(pageNum int, pageSize int, maps interface{}) (count int64, vuls []Vul) {
	globalDBTmp := GlobalDB.Model(&Vul{})
	query := maps.(map[string]interface{})

	if query["project"] != nil {
		globalDBTmp = globalDBTmp.Where("project LIKE ?", "%"+query["project"].(string)+"%")
	}

	if query["rule_id"] != nil {
		globalDBTmp = globalDBTmp.Where("rule_id LIKE ?", "%"+query["rule_id"].(string)+"%")
	}

	globalDBTmp.Where("handled = ?", false).Count(&count)

	globalDBTmp.Offset(pageNum).Limit(pageSize).Order("id asc").Find(&vuls)

	return
}

func VulTotal() (count int64) {
	globalDBTmp := GlobalDB.Model(&Vul{})
	globalDBTmp.Count(&count)

	return
}

func DeleteVul(id string) {
	globalDBTmp := GlobalDB.Model(&Vul{})
	globalDBTmp.Where("id = ?", id).Unscoped().Delete(&Vul{})
}

// ExistVul  Determine whether ip and port exist in the database
func ExistVul(id string) (bool, Vul) {
	var vul Vul
	globalDBTmp := GlobalDB.Model(&Vul{})
	globalDBTmp.Where("id = ?", id).Limit(1).First(&vul)

	if vul.Id > 0 {
		return true, vul
	}

	return false, vul
}

// UpdateHandled Update Fields
func UpdateHandled(id string) {
	globalDBTmp := GlobalDB.Model(&Vul{})
	globalDBTmp.Where("id = ?", id).Update("handled", true)
}
