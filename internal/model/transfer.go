package model

import "gorm.io/gorm"

type Transfer struct {
	gorm.Model
	SourAddr   string  `json:"sour_addr"`
	DescAddr   string  `json:"desc_addr"`
	Amount     float64 `json:"amount"`
	Signatures string  `json:"signatures"`
	Data       string  `json:"data"`
	TransferTime string `json:"transfer_time"`
}

func (m *Transfer) TableName() string {
	return "transfer"
}
