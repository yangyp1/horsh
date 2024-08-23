package model

import "gorm.io/gorm"

type Invitations struct {
	gorm.Model
	InviterID string `json:"inviter_id"` //邀请人ID
	InviteeID string `json:"invitee_id"` //被邀请人ID
	Level     int    `json:"level"`      //邀请层级
}

func (m *Invitations) TableName() string {
	return "invitations"
}
