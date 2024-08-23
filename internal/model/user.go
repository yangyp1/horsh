package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserId            string  `gorm:"unique;not null" json:"user_id"` //用户ID
	Address           string  `gorm:"not null" json:"address"`        //Sol地址
	EvmAddress        string  `json:"evm_address"`                    //Evm地址
	Code              string  `gorm:"not null" json:"code"`           //邀请码
	InviteBy          string  `gorm:"not null" json:"invite_by"`      //邀请人ID
	SolAmount         float64 `json:"sol_amount"`                     //捐赠sol数量
	Nodes             int     `json:"nodes"`                          //节点数量
	HorshCount        int     `json:"horsh_count"`
	UsdtCount         int     `json:"usdt_count"`
	DirectReward      int     `json:"direct_reward"`       //直邀奖励
	ClaimDirectReward int     `json:"claim_direct_reward"` //提取直邀奖励
	TeamReward        int     `json:"team_reward"`         //团邀奖励
	ClaimTeamReward   int     `json:"claim_team_reward"`   //提取团邀奖励
}

func (u *User) TableName() string {
	return "users"
}
