package model

import "gorm.io/gorm"

type Teamcounts struct {
	gorm.Model
	UserId    string `json:"user_id"`    //用户ID
	TeamCount int    `json:"team_count"` //团队人数
}

func (m *Teamcounts) TableName() string {
	return "teamcounts"
}
