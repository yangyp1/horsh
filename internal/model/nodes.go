package model

import "gorm.io/gorm"

type Nodes struct {
	gorm.Model
	UserId      string  `json:"user_id"`      // 用户id
	Address     string  `json:"address"`      // 购买地址
	RecvAddress string  `json:"recv_address"` // 收款地址
	Amount      float64 `json:"amount"`       //数量
	ProgramId   string  `json:"program_id"`
	Signatures  string  `json:"signatures"`
}

func (m *Nodes) TableName() string {
	return "nodes"
}
