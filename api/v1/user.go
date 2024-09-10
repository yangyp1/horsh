package v1

type RegisterRequest struct {
	Address      string `json:"address" binding:"required" example:"0x2w323423"`
	OriginalMsg  string `json:"original_msg" binding:"required"`
	SignatureHex string `json:"signature_hex" binding:"required"`
}

type LoginRequest struct {
	Address      string `json:"address" binding:"required" example:"0x2w323423"`
	OriginalMsg  string `json:"original_msg" binding:"required"`
	SignatureHex string `json:"signature_hex" binding:"required"`
	LoginType    string `json:"login_type" binding:"required,oneof=sol evm"`
}

type ClaimRequest struct {
	Count     int    `json:"count" binding:"required"`                        //数量
	ClaimType string `json:"claim_type" binding:"required,oneof=direct team"` //直推奖励，团队奖励
}

//

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required" example:"0x2w323423"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseData struct {
	AccessToken string  `json:"accessToken"` //token
	Address     string  `json:"address"`     //SOL钱包地址
	UserId      string  `json:"User_id"`     //用户ID
	Code        string  `json:"code"`        //自己的邀请码
	SolAmount   float64 `json:"sol_amount"`  //捐赠SOL数量
	EvmAddress  string  `json:"evm_address"` //evm钱包地址
	HorshCount  int     `json:"horsh_count"`
	UsdtCount   int     `json:"usdt_count"`
}
type LoginResponse struct {
	Response
	Data LoginResponseData
}

type InviteCodeRequest struct {
	UserId string `json:"user_id" binding:"required"`
	Code   string `json:"Code" binding:"required"`
}
type InviteCodeResponseData struct {
	Code   string `json:"accessToken"` //token
	UserId string `json:"User_id"`     //用户ID
}
type InviteCodeResponse struct {
	Response
	Data InviteCodeResponseData
}

type SelectResponseData struct {
	EvmAddress        string  `json:"evm_address"`         //EVM地址
	GlobalNodes       int     `json:"global_nodes"`        //节点总数
	MyNodes           int     `json:"my_nodes"`            //我的节点数量
	EstimatedEarnings int     `json:"estimated_earnings"`  //预计收益
	MyStable          int     `json:"my_stable"`           //我的马场
	MyRacehorses      int     `json:"my_racehorses"`       //我的赛马
	DirectPerformance int     `json:"direct_performance"`  //直邀业绩
	DirectReward      int     `json:"direct_reward"`       //直邀奖励
	ClaimDirectReward int     `json:"claim_direct_reward"` //提取直邀奖励
	TeamReward        int     `json:"team_reward"`         //团邀奖励
	ClaimTeamReward   int     `json:"claim_team_reward"`   //提取团邀奖励
	TeamPerformance   int     `json:"team_performance"`    //团队业绩
	DirectCount       int64   `json:"direct_count"`        //直推人数
	TeamCount         int64   `json:"team_count"`          //团队人数
	DirectSolAmount   string  `json:"direct_sol_amount"`   //直推捐款sol数量
	TeamSolAmount     string  `json:"team_sol_amount"`     //团队捐款sol数量
	Bonus             float64 `json:"bonus"`               //马头奖励

}
type SelectResponse struct {
	Response
	Data SelectResponseData
}

type SolSearchRequest struct {
	Address string `json:"address"` //Evm地址
}

type AdminSearchRequest struct {
	Address          string `json:"address"`                                  //地址
	Code             string `json:"code"`                                     //邀请码
	StartTime        string `json:"start_time" example:"2024-06-01 15:03:04"` //开始时间
	EndTime          string `json:"end_time" example:"2024-07-01 15:03:04"`   //结束时间
	IsInviteby       bool   `json:"is_inviteby"`                              //是否去除被邀请
	InviteBy         string `json:"invite_by"`                                //被谁邀请
	IsNotEmptyAmount bool   `json:"is_not_empty_amount"`                      //SOL数量不为空
	Page             int    `json:"page"`                                     //"分页"
	PageSize         int    `json:"page_size"`                                // "每页数量"
}

type ExportRequest struct {
	Address   string `json:"address"`                                  //地址
	StartTime string `json:"start_time" example:"2024-06-01 15:03:04"` //开始时间
	EndTime   string `json:"end_time" example:"2024-07-01 15:03:04"`   //结束时间
}

type ExportTeamRequest struct {
	UserId string `json:"user_id"` //地址
}

type ExportResponse struct {
	SourAddr     string  `json:"sour_addr"`
	Amount       float64 `json:"amount"`
	TransferTime string  `json:"transfer_time"`
	ID           uint    `json:"id"`
	UserId       string  `json:"user_id"`
	Code         string  `json:"code"`
	InviteBy     string  `json:"invite_by"`
}

type ExportResponsezTeam struct {
	Address    string  `json:"address"`
	Sol_amount float64 `json:"sol_amount"`
	CreatedAt  string  `json:"created_at"`
	ID         uint    `json:"id"`
	UserId     string  `json:"user_id"`
	Code       string  `json:"code"`
	InviteBy   string  `json:"invite_by"`
}

type ExportResponseData struct {
	UserId          string  `gorm:"unique;not null" json:"user_id" csv:"用户id"` //用户id
	Address         string  `gorm:"not null" json:"address" csv:"地址"`          //地址
	Code            string  `gorm:"not null" json:"code" csv:"邀请码"`            //邀请码
	InviteBy        string  `gorm:"not null" json:"invite_by" csv:"上级ID"`      //上级ID
	SolAmount       float64 `json:"sol_amount" csv:"SOL"`                      //本人SOL
	DirectCount     int64   `json:"direct_count" csv:"直推人数"`                   //直推人数
	TeamCount       int64   `json:"team_count" csv:"团队人数"`                     //团队人数
	DirectSolAmount string  `json:"direct_sol_amount" csv:"直推sol数量"`           //直推sol数量
	TeamSolAmount   string  `json:"team_sol_amount" csv:"团队sol数量"`             //团队sol数量
	BonusTime       string  `json:"bonus_time" csv:"打款时间"`                     //捐赠时间
	Bonus           string  `json:"bonus", csv:"bonus"`                        //马头
}

type AdminSearchResponseData struct {
	UserId          string  `gorm:"unique;not null" json:"user_id" csv:"用户id"` //用户id
	Address         string  `gorm:"not null" json:"address" csv:"地址"`          //地址
	EvmAddress      string  `json:"evm_address" csv:"EVM地址"`
	Code            string  `gorm:"not null" json:"code" csv:"邀请码"`       //邀请码
	InviteBy        string  `gorm:"not null" json:"invite_by" csv:"上级ID"` //上级ID
	//SolAmount       float64 `json:"sol_amount" csv:"本人SOL"`               //本人SOL
	DirectCount     int64   `json:"direct_count" csv:"直推人数"`              //直推人数
	TeamCount       int64   `json:"team_count" csv:"团队人数"`                //团队人数
	PeosonalNode    int64   `json:"peosonal_node" csv:"个人入金"`             //个人入金
	DirectNode      int64   `json:"direct_node" csv:"直推入金"`               //直推入金
	TeamNode        int64   `json:"team_node" csv:"团队入金"`                 //团队入金
	//DirectSolAmount string  `json:"direct_sol_amount" csv:"直推捐款sol数量"`    //直推捐款sol数量
	//TeamSolAmount   string  `json:"team_sol_amount" csv:"团队捐款sol数量"`      //团队捐款sol数量
	//BonusTime       string  `json:"bonus_time" csv:"捐赠时间"`                //捐赠时间
	//Bonus           string  `json:"bonus" csv:"个人捐赠马头"`                   //个人捐赠马头
}

type AdminSearchResponseDa struct {
	Count      int64                     //数量
	ReturnData []AdminSearchResponseData //数据
}

type AdminSearchResponse struct {
	Response
	Data AdminSearchResponseDa
}

type SolSearchResponData struct {
	UserId            string `json:"user_id"`
	EvmAddress        string `json:"evm_address"`
	DirectCount       int64  `json:"direct_count"`       //直推人数
	TeamCount         int64  `json:"team_count"`         //团队人数
	DirectPerformance int    `json:"direct_performance"` //直邀业绩
	TeamPerformance   int    `json:"team_performance"`   //团队业绩
}
type SolSearchResponse struct {
	Response
	Data SolSearchResponData
}

type BindEvmAddressRequest struct {
	EvmAddress string `json:"evm_address" example:"0x0324adsdacfdtswcsdds23eecwr"`
}

type HorshTransferRequest struct {
	HorshHash string `json:"horsh_hash"  binding:"required"`
}

type USDTTransferRequest struct {
	UsdtHash string `json:"usdt_hash"  binding:"required"`
}

type GetProfileResponseData struct {
	UserId string `json:"userId"`
}
type GetProfileResponse struct {
	Response
	Data GetProfileResponseData
}

type Invitecount struct {
	TeamCount   int64 `json:"teamcount" gorm:"column:teamcount"`
	DirectCount int64 `json:"directcount" gorm:"column:directcount"`
}

type InviteAmount struct {
	DirectAmount float64 `json:"directamount" gorm:"column:directamount"`
	TeamAmount   float64 `json:"teamamount" gorm:"column:teamamount"`
}


type InviteNode struct {
	DirectNode float64 `json:"directnode" gorm:"column:directnode"`
	TeamNode   float64 `json:"teamnode" gorm:"column:teamnode"`
}

type AdminAllCountResponseData struct {
	RegisCount  int     `json:"regis_count" gorm:"column:regis_count"`
	BonusCount  int     `json:"bonus_count" gorm:"column:bonus_count"`
	TotolAmount float64 `json:"totol_amount" gorm:"column:totol_amount"`
}

type AllCountResponse struct {
	Response
	Data AdminAllCountResponseData
}
