package service

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/model"
	"SolProject/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/tokenprog"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

type Transaction struct {
	Result TransactionData `json:"result"`
}

type TransactionData struct {
	BlockTime   float64 `json:"blockTime"`
	Meta        Meta    `json:"meta"`
	Slot        float64 `json:"slot"`
	Transaction Txn     `json:"transaction"`
}
type Meta struct {
	Err          interface{} `json:"err"`
	PostBalances []float64   `json:"postBalances"`
	PreBalances  []float64   `json:"preBalances"`
}
type Txn struct {
	Message    Message  `json:"message"`
	Signatures []string `json:"signatures"`
}
type Message struct {
	AccountKeys  []string      `json:"accountKeys"`
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	ProgramIdIndex int    `json:"programIdIndex"`
}

type UserService interface {
	Login(ctx context.Context, req *v1.LoginRequest) (*model.User, string, error)
	AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (string, error)
	AdminSearch(ctx context.Context, req *v1.AdminSearchRequest, limit int, offset int) (int64, *[]v1.AdminSearchResponseData, error)
	GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error)
	BindEvmAddress(ctx context.Context, userId string, req v1.BindEvmAddressRequest) error
	Select(ctx context.Context, user_id string) (*v1.SelectResponseData, error)
	InviteCode(ctx context.Context, req *v1.InviteCodeRequest) (string, error)
	Solsearch(ctx context.Context, user_id string, address string) (*[]v1.SolSearchResponData, error)
	AdminAllCount(ctx context.Context) (*v1.AdminAllCountResponseData, error)
	CreationCount(ctx context.Context, user_id string) (int, error)
	ExportRecord(ctx context.Context, req *v1.ExportRequest) (*[]v1.ExportResponseData, error)
	ExportRecordTeam(ctx context.Context, req *v1.ExportTeamRequest) (*[]v1.ExportResponseData, error)
	ExportRecordRegis(ctx context.Context) (*[]v1.ExportResponseData, error)
	HorshTransfer(ctx context.Context, userId string, req v1.HorshTransferRequest) (*model.User, int, error)
	USDTTransfer(ctx context.Context, userId string, req v1.USDTTransferRequest) (*model.User, int, error)
	ClaimReward(ctx context.Context, userId string, req v1.ClaimRequest) (*model.User, error)
	TASK(ctx context.Context)
}

func NewUserService(
	service *Service,
	conf *viper.Viper,
	userRepo repository.UserRepository,
	inviteRepo repository.InvitationsRepository,
	teamRepo repository.TeamcountsRepository,
	trsfRepo repository.TransferRepository,
	nodesRepo repository.NodesRepository) UserService {
	return &userService{
		userRepo:   userRepo,
		conf:       conf,
		inviteRepo: inviteRepo,
		teamRepo:   teamRepo,
		Service:    service,
		trsfRepo:   trsfRepo,
		nodesRepo:  nodesRepo,
	}
}

type userService struct {
	userRepo   repository.UserRepository
	conf       *viper.Viper
	inviteRepo repository.InvitationsRepository
	teamRepo   repository.TeamcountsRepository
	trsfRepo   repository.TransferRepository
	nodesRepo  repository.NodesRepository
	*Service
}

// 设置 Solana RPC 端点
const SOLANA_RPC_URL = "https://api.mainnet-beta.solana.com"

// 获取交易签名列表
func getSignatures(address string, lastSignature string) (map[string]interface{}, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSignaturesForAddress",
		"params": []interface{}{
			address,
			map[string]interface{}{"limit": 1000, "until": lastSignature},
		},
	}
	return postRequest(SOLANA_RPC_URL, headers, payload)
}

// 获取交易详情
func getTransaction(signature string) (Transaction, error) {
	returndata := Transaction{}
	client := resty.New()
	client.Debug = true
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getTransaction",
			"params":  []interface{}{signature},
		}).Post(SOLANA_RPC_URL)
	if err != nil {
		return returndata, err
	}
	if response.StatusCode() != http.StatusOK {
		return returndata, fmt.Errorf("code=%d,response=%s", response.StatusCode(), response.String())
	}
	err = json.Unmarshal(response.Body(), &returndata)
	if err != nil {
		return returndata, err
	}
	return returndata, nil
}

// 发送POST请求
func postRequest(url string, headers map[string]string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 解析交易数据
func parseTransactions(signature string, targetAddress string) ([]model.Transfer, error) {
	receivedTransactions := []model.Transfer{}
	txDetail := Transaction{}
	client := resty.New()
	client.Debug = true
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getTransaction",
			"params":  []interface{}{signature},
		}).Post(SOLANA_RPC_URL)
	if err != nil {
		return receivedTransactions, err
	}
	if response.StatusCode() != http.StatusOK {
		return receivedTransactions, fmt.Errorf("code=%d,err=%s", response.StatusCode(), response.String())
	}
	err = json.Unmarshal(response.Body(), &txDetail)
	if err != nil {
		return receivedTransactions, err
	}
	//fmt.Printf("signatureprint txDetail[result] Err:%#v\n", txDetail.Result.Meta.Err)
	//fmt.Printf("signatureprint txDetail[result] AccountKeys:%#v\n", txDetail.Result.Transaction.Message.AccountKeys)
	//fmt.Printf("signatureprint txDetail[result] Instructions:%#v\n", txDetail.Result.Transaction.Message.Instructions)
	blocktime := txDetail.Result.BlockTime
	transfertime := time.Unix(int64(blocktime), 0)
	transaction := txDetail.Result.Transaction
	meta := txDetail.Result.Meta
	if meta.Err == nil {
		for _, inst := range transaction.Message.Instructions {
			if len(inst.Accounts) == 2 {
				data := inst.Data
				amount := (meta.PostBalances[inst.Accounts[1]] - meta.PreBalances[inst.Accounts[1]]) / 1000000000
				sourAddr := transaction.Message.AccountKeys[inst.Accounts[0]]
				destAddr := transaction.Message.AccountKeys[inst.Accounts[1]]
				if destAddr == targetAddress {
					receivedTransactions = append(receivedTransactions, model.Transfer{
						SourAddr:     sourAddr,
						DescAddr:     destAddr,
						Data:         data,
						Signatures:   signature,
						Amount:       amount,
						TransferTime: transfertime.Format("2006-01-02 15:04:05"),
					})
				}
			}
		}
	}
	return receivedTransactions, nil
}

func (s *userService) TASK(ctx context.Context) {

	targetAddress := s.conf.GetString("data.sol_address") //"6z3kAFzN9rZ14h5qfTmEk4VvCepr1tVot6uXxWxjdU6K"
	lastSignature := "4t2RWskCSemC1VDWDhpCfZU3rmdGCTL2ASJBcwUT5T1hqer39u7nMhu8vtKoK9AeqPRCLtL8Y1fhB81Dnm3cbLU5"
	for {
		//fmt.Printf("signatureprint INFO lastSignature:%s\n", lastSignature)
		signaturesResponse, err := getSignatures(targetAddress, lastSignature)
		if err != nil {
			fmt.Println("signatureprint getSignatures Error:", err)
			time.Sleep(30 * time.Second) // 如果查询失败，等待10秒后重试
			continue
		}
		fmt.Printf("signatureprint INFO signaturesResponse:%s\n", signaturesResponse)
		if result, ok := signaturesResponse["result"]; ok {
			transactions := result.([]interface{})
			fmt.Printf("signatureprint INFO:%d\n", len(transactions))
			for i := len(transactions) - 1; i >= 0; i-- {
				tx := transactions[i]
				txMap := tx.(map[string]interface{})
				signature := txMap["signature"].(string)
				fmt.Printf("signatureprint INFO:%s\n", signature)
				receivedTransactions, err := parseTransactions(signature, targetAddress)
				if err != nil {
					fmt.Println("signatureprint parseTransactions Error:", err)
					time.Sleep(30 * time.Second) // 如果查询失败，等待10秒后重试
					continue
				}
				for _, transfer := range receivedTransactions {
					err = s.tm.Transaction(ctx, func(ctx context.Context) error {
						// 创建转账记录
						if err = s.trsfRepo.Create(ctx, &transfer); err != nil {
							return err
						}
						// 更新用户捐赠数量
						if err := s.userRepo.UpdateColumn(ctx, transfer.SourAddr, transfer.Amount); err != nil {
							return err
						}
						return nil
					})
					if err != nil {
						fmt.Printf("signatureprint UPDATE treanfer err :%s\n", err)
					}

				}
				lastSignature = signature
				time.Sleep(5 * time.Second)

			}
		}
		time.Sleep(30 * time.Second)
	}
}

func (s *userService) Login(ctx context.Context, req *v1.LoginRequest) (*model.User, string, error) {
	if req.LoginType == "evm" {
		evmuser, err := s.userRepo.GetByEvmAddress(ctx, req.Address)
		if err != nil {
			s.logger.WithContext(ctx).Sugar().Errorf("GetByEvmAddress err = %s", err.Error())
			return nil, "", err
		}
		if evmuser == nil {
			return nil, "", fmt.Errorf("EvmAddress Not Exist")
		}
		token, err := s.jwt.GenToken(evmuser.UserId, time.Now().Add(time.Hour*24*90))
		if err != nil {
			return evmuser, "", err
		}
		return evmuser, token, nil
	}
	user, err := s.userRepo.GetByAddress(ctx, req.Address)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAddress err = %s", err.Error())
		return user, "", v1.ErrInternalServerError
	}
	if user != nil {
		token, err := s.jwt.GenToken(user.UserId, time.Now().Add(time.Hour*24*90))
		if err != nil {
			return user, "", err
		}
		return user, token, nil
	}
	userId, err := s.sid.GenString()
	if err != nil {
		return user, "", err
	}
	newUser := &model.User{
		UserId:  userId,
		Address: req.Address,
	}
	err = s.userRepo.Create(ctx, newUser)
	if err != nil {
		return user, "", err
	}
	token, err := s.jwt.GenToken(newUser.UserId, time.Now().Add(time.Hour*24*90))
	if err != nil {
		return user, "", err
	}
	return newUser, token, nil

}

func (s *userService) AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (string, error) {
	token, err := s.jwt.GenToken(req.Username, time.Now().Add(time.Hour*24*90))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) AdminSearch(ctx context.Context, req *v1.AdminSearchRequest, limit int, offset int) (int64, *[]v1.AdminSearchResponseData, error) {
	repos := []v1.AdminSearchResponseData{}
	count, users, err := s.userRepo.GetByAll(ctx, req, limit, offset)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAddress err = %s", err.Error())
		return count, &repos, v1.ErrInternalServerError
	}
	for _, user := range *users {
		invicount, err := s.inviteRepo.GetInviteCount(ctx, user.UserId)
		if err != nil {
			return count, &repos, err
		}
		node, err := s.inviteRepo.GetInviteNode(ctx, user.UserId)
		if err != nil {
			return count, &repos, err
		}
		repos = append(repos, v1.AdminSearchResponseData{
			UserId:       user.UserId,
			Address:      user.Address,
			EvmAddress:   user.EvmAddress,
			Code:         user.Code,
			InviteBy:     user.InviteBy,
			//SolAmount:    user.SolAmount,
			DirectCount:  invicount.DirectCount,
			TeamCount:    invicount.TeamCount,
			PeosonalNode: int64(user.Nodes),
			DirectNode:   int64(node.DirectNode),
			TeamNode:     int64(node.TeamNode),
			//DirectSolAmount: fmt.Sprintf("%.2f", amount.DirectAmount),
			//TeamSolAmount:   fmt.Sprintf("%.2f", amount.TeamAmount),
			//BonusTime:       user.UpdatedAt.Format("2006-01-02 15:04:05"),
			//Bonus:           fmt.Sprintf("%.4f", amount.DirectAmount*0.05),
		})

	}
	return count, &repos, nil
}

func (s *userService) ExportRecord(ctx context.Context, req *v1.ExportRequest) (*[]v1.ExportResponseData, error) {
	repos := []v1.ExportResponseData{}
	transfers, err := s.trsfRepo.GetByAll(ctx, req)

	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAddress err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	for _, taansfer := range *transfers {

		invicount, err := s.inviteRepo.GetInviteCount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}

		amount, err := s.inviteRepo.GetInviteAmount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}
		repos = append(repos, v1.ExportResponseData{
			UserId:          taansfer.UserId,
			Address:         taansfer.SourAddr,
			Code:            taansfer.Code,
			InviteBy:        taansfer.InviteBy,
			SolAmount:       taansfer.Amount,
			DirectCount:     invicount.DirectCount,
			TeamCount:       invicount.TeamCount,
			DirectSolAmount: fmt.Sprintf("%.2f", amount.DirectAmount),
			TeamSolAmount:   fmt.Sprintf("%.2f", amount.TeamAmount),
			BonusTime:       taansfer.TransferTime,
			Bonus:           fmt.Sprintf("%.4f", amount.DirectAmount*0.05),
		})

	}
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAll err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	return &repos, nil
}

func (s *userService) ExportRecordRegis(ctx context.Context) (*[]v1.ExportResponseData, error) {
	repos := []v1.ExportResponseData{}
	transfers, err := s.trsfRepo.GetByAllRegis(ctx)

	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAddress err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	for _, taansfer := range *transfers {

		invicount, err := s.inviteRepo.GetInviteCount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}

		amount, err := s.inviteRepo.GetInviteAmount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}
		repos = append(repos, v1.ExportResponseData{
			UserId:          taansfer.UserId,
			Address:         taansfer.Address,
			Code:            taansfer.Code,
			InviteBy:        taansfer.InviteBy,
			SolAmount:       taansfer.Sol_amount,
			DirectCount:     invicount.DirectCount,
			TeamCount:       invicount.TeamCount,
			DirectSolAmount: fmt.Sprintf("%.2f", amount.DirectAmount),
			TeamSolAmount:   fmt.Sprintf("%.2f", amount.TeamAmount),
			BonusTime:       taansfer.CreatedAt,
			Bonus:           fmt.Sprintf("%.4f", amount.DirectAmount*0.05),
		})

	}
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAll err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	return &repos, nil
}

func (s *userService) ExportRecordTeam(ctx context.Context, req *v1.ExportTeamRequest) (*[]v1.ExportResponseData, error) {
	repos := []v1.ExportResponseData{}
	transfers, err := s.trsfRepo.GetByAllTeam(ctx, req)

	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAddress err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	for _, taansfer := range *transfers {

		invicount, err := s.inviteRepo.GetInviteCount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}

		amount, err := s.inviteRepo.GetInviteAmount(ctx, taansfer.UserId)
		if err != nil {
			return &repos, err
		}
		repos = append(repos, v1.ExportResponseData{
			UserId:          taansfer.UserId,
			Address:         taansfer.Address,
			Code:            taansfer.Code,
			InviteBy:        taansfer.InviteBy,
			SolAmount:       taansfer.Sol_amount,
			DirectCount:     invicount.DirectCount,
			TeamCount:       invicount.TeamCount,
			DirectSolAmount: fmt.Sprintf("%.2f", amount.DirectAmount),
			TeamSolAmount:   fmt.Sprintf("%.2f", amount.TeamAmount),
			BonusTime:       taansfer.CreatedAt,
			Bonus:           fmt.Sprintf("%.4f", amount.DirectAmount*0.05),
		})

	}
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByAll err = %s", err.Error())
		return &repos, v1.ErrInternalServerError
	}
	return &repos, nil
}

func (s *userService) AdminAllCount(ctx context.Context) (*v1.AdminAllCountResponseData, error) {
	allcount, err := s.userRepo.GetAllCounty(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetAllCount err = %s", err.Error())
		return allcount, v1.ErrInternalServerError
	}

	return allcount, nil
}

func (s *userService) CreationCount(ctx context.Context, code string) (int, error) {
	user, err := s.userRepo.GetByCode(ctx, code)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, v1.ErrUserNotExist
	}
	invitecodes := []string{user.UserId} //ByYqw7hii
	allcount := 0
	for {
		inviteeids, err := s.inviteRepo.CreationCount(ctx, invitecodes)
		if err != nil {
			s.logger.WithContext(ctx).Sugar().Errorf("GetAllCount err = %s", err.Error())
			return allcount, v1.ErrInternalServerError
		}
		if len(inviteeids) == 0 {
			break
		}
		invitecodes = inviteeids
		allcount += len(inviteeids)
	}
	return allcount, nil
}

type TransactionTranfser struct {
	Result TransactionDataTrans `json:"result"`
}

type TransactionDataTrans struct {
	BlockTime   float64   `json:"blockTime"`
	Meta        TransMeta `json:"meta"`
	Slot        float64   `json:"slot"`
	Transaction Txn       `json:"transaction"`
}
type TransMeta struct {
	Err               interface{}     `json:"err"`
	PostBalances      []float64       `json:"postBalances"`
	PreBalances       []float64       `json:"preBalances"`
	PreTokenBalances  []TokenBalances `json:"preTokenBalances"`
	PostTokenBalances []TokenBalances `json:"postTokenBalances"`
}

type TokenBalances struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner"`
	ProgramId     string        `json:"programId"`
	UiTokenAmount UiTokenAmount `json:"uiTokenAmount"`
}
type UiTokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}

func GetTransaction(signature string) (TransactionTranfser, error) {
	returndata := TransactionTranfser{}
	client := resty.New()
	client.Debug = true
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getTransaction",
			"params":  []interface{}{signature},
		}).Post("https://api.devnet.solana.com")
	if err != nil {
		return returndata, err
	}
	if response.StatusCode() != http.StatusOK {
		return returndata, fmt.Errorf("code=%d,response=%s", response.StatusCode(), response.String())
	}
	err = json.Unmarshal(response.Body(), &returndata)
	if err != nil {
		return returndata, err
	}
	return returndata, nil
}

func (s *userService) ClaimReward(ctx context.Context, userId string, req v1.ClaimRequest) (*model.User, error) {
	user, err := s.userRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, v1.ErrUserNotExist
	}
	if req.ClaimType == "direct" && req.Count > user.DirectReward {
		return nil, fmt.Errorf("user reward not enough")
	}
	if req.ClaimType == "team" && req.Count > user.TeamReward {
		return nil, fmt.Errorf("user reward not enough")
	}
	// 初始化客户端
	c := client.NewClient(rpc.DevnetRPCEndpoint)
	// 发送者钱包
	senderPrivateKey := "gz9nNevUm3UYC1jnbjshVkk2GXbcLD51TqQySWSeHxytLHJoMF9uxEBQNGXcvYgZMpANH6UJdHAEcRhWSiCFFcj"
	sender, err := types.AccountFromBase58(senderPrivateKey)
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	// 接收者钱包地址
	receiverPubKey := user.Address
	// SPL 代币的 Token Mint Address
	tokenMintAddress := "BxdTuhEe2cQnr58ffZVe8VMXF7cmBH2PXgwNCA7WvfFz"
	// 获取或创建发送者的 SPL 代币账户地址
	senderTokenAccount, _, err := common.FindAssociatedTokenAddress(sender.PublicKey, common.PublicKeyFromString(tokenMintAddress))
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	// 获取或创建接收者的 SPL 代币账户地址
	receiverTokenAccount, _, err := common.FindAssociatedTokenAddress(common.PublicKeyFromString(receiverPubKey), common.PublicKeyFromString(tokenMintAddress))
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	fmt.Printf("receiverTokenAccount = %s\n", receiverTokenAccount.String())
	// 转账 SPL 代币
	recentBlockhash, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{sender},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        sender.PublicKey,
			RecentBlockhash: recentBlockhash.Blockhash,
			Instructions: []types.Instruction{
				tokenprog.TransferChecked(tokenprog.TransferCheckedParam{
					From:     senderTokenAccount,
					To:       receiverTokenAccount,
					Mint:     common.PublicKeyFromString(tokenMintAddress),
					Auth:     sender.PublicKey,
					Amount:   uint64(req.Count) * 100000000, // 转账数量（以最小单位为单位）
					Decimals: 8,                             // SPL 代币的小数位数
				}),
			},
		}),
	})
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	// 签名并发送交易
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		s.logger.Sugar().Infof("err = %s", err.Error())
		return nil, err
	}
	if req.ClaimType == "direct" {
		user.DirectReward = user.DirectReward - req.Count
		user.ClaimDirectReward = user.ClaimDirectReward + req.Count
	} else if req.ClaimType == "team" {
		user.TeamReward = user.TeamReward - req.Count
		user.ClaimTeamReward = user.ClaimTeamReward + req.Count
	}
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	fmt.Println("Transaction sent:", txhash)
	return user, nil

}

func (s *userService) HorshTransfer(ctx context.Context, userId string, req v1.HorshTransferRequest) (*model.User, int, error) {
	// 查询交易详情
	Horsh := model.Nodes{}
	tx, err := GetTransaction(req.HorshHash)
	if err != nil {
		log.Fatalf("Failed to get transaction details: %v", err)
		return &model.User{}, 0, fmt.Errorf("Failed to get transaction details: %v", err)
	}
	user, _ := s.userRepo.GetByUserId(ctx, userId)
	if user == nil {
		return &model.User{}, 0, fmt.Errorf("user not exist")
	}
	// 解析交易中的指令
	s.logger.Sugar().Infof("%#v", tx.Result)
	confirm_status := 0
	if len(tx.Result.Transaction.Message.AccountKeys) == 0 {
		return user, 0, nil
	}
	Horsh.Address = tx.Result.Transaction.Message.AccountKeys[0]
	for i, postBalance := range tx.Result.Meta.PostTokenBalances {
		preBalance := tx.Result.Meta.PreTokenBalances[i]
		if postBalance.UiTokenAmount.UiAmount > preBalance.UiTokenAmount.UiAmount {
			user.HorshCount += 1
			min := func() int {
				if user.HorshCount > user.UsdtCount {
					return user.UsdtCount
				}
				return user.HorshCount
			}()
			confirm_status = 1
			Horsh.UserId = userId
			Horsh.RecvAddress = postBalance.Owner
			Horsh.Amount = postBalance.UiTokenAmount.UiAmount - preBalance.UiTokenAmount.UiAmount
			Horsh.ProgramId = postBalance.ProgramId
			Horsh.Signatures = tx.Result.Transaction.Signatures[0]
			err = s.tm.Transaction(ctx, func(ctx context.Context) error {
				// 创建用户
				if err = s.userRepo.UpDateByHorshCount(ctx, userId); err != nil {
					return err
				}
				//插入交易记录
				if err = s.nodesRepo.Create(ctx, &Horsh); err != nil {
					return err
				}
				if min > user.Nodes {
					//修改用户节点信息
					user.Nodes = min
					if err = s.userRepo.Update(ctx, user); err != nil {
						return err
					}
					//增加直接邀请奖励
					directUser, err := s.userRepo.GetByUserId(ctx, user.InviteBy)
					if err != nil {
						return err
					}
					if directUser != nil {
						directUser.DirectReward = directUser.DirectReward + 50
						if err = s.userRepo.Update(ctx, directUser); err != nil {
							return err
						}
					}
					//增加间接邀请奖励
					invitations, err := s.inviteRepo.FindByInviteeId(ctx, user.UserId, 16)
					if err != nil {
						return err
					}
					for _, invitation := range *invitations {
						if err = s.userRepo.UpDateByTeamReward(ctx, invitation.InviterID); err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return &model.User{}, 0, err
			}
			break
		}
	}
	return user, confirm_status, nil

}

func (s *userService) USDTTransfer(ctx context.Context, userId string, req v1.USDTTransferRequest) (*model.User, int, error) {
	// 查询交易详情
	USDT := model.Nodes{}
	tx, err := GetTransaction(req.UsdtHash)
	if err != nil {
		log.Fatalf("Failed to get transaction details: %v", err)
		return &model.User{}, 0, fmt.Errorf("Failed to get transaction details: %v", err)
	}
	user, _ := s.userRepo.GetByUserId(ctx, userId)
	if user == nil {
		return &model.User{}, 0, fmt.Errorf("user not exist")
	}
	if len(tx.Result.Transaction.Message.AccountKeys) == 0 {
		return user, 0, nil
	}
	// 解析交易中的指令
	confirm_status := 0
	s.logger.Sugar().Infof("%#v", tx.Result)
	USDT.Address = tx.Result.Transaction.Message.AccountKeys[0]
	for i, postBalance := range tx.Result.Meta.PostTokenBalances {
		preBalance := tx.Result.Meta.PreTokenBalances[i]
		if postBalance.UiTokenAmount.UiAmount > preBalance.UiTokenAmount.UiAmount {
			user.UsdtCount += 1
			min := func() int {
				if user.HorshCount > user.UsdtCount {
					return user.UsdtCount
				}
				return user.HorshCount
			}()
			confirm_status = 1
			USDT.UserId = userId
			USDT.RecvAddress = postBalance.Owner
			USDT.Amount = postBalance.UiTokenAmount.UiAmount - preBalance.UiTokenAmount.UiAmount
			USDT.ProgramId = postBalance.ProgramId
			USDT.Signatures = tx.Result.Transaction.Signatures[0]
			err = s.tm.Transaction(ctx, func(ctx context.Context) error {
				// 更新usdt节点
				if err = s.userRepo.UpDateByUSDTCount(ctx, userId); err != nil {
					return err
				}
				//插入交易记录
				if err = s.nodesRepo.Create(ctx, &USDT); err != nil {
					return err
				}
				if min > user.Nodes {
					// 修改用户节点信息
					user.Nodes = min
					if err = s.userRepo.Update(ctx, user); err != nil {
						return err
					}
					//增加直接邀请奖励
					directUser, err := s.userRepo.GetByUserId(ctx, user.InviteBy)
					if err != nil {
						return err
					}
					if directUser != nil {
						directUser.DirectReward = directUser.DirectReward + 50
						if err = s.userRepo.Update(ctx, directUser); err != nil {
							return err
						}
					}
					//增加间接邀请奖励
					// 插入间接邀请关系
					invitations, err := s.inviteRepo.FindByInviteeId(ctx, user.UserId, 15)
					if err != nil {
						return err
					}
					for _, invitation := range *invitations {
						if err = s.userRepo.UpDateByTeamReward(ctx, invitation.InviterID); err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return &model.User{}, 0, err
			}
			break
		}
	}
	return user, confirm_status, nil
}

func (s *userService) InviteCode(ctx context.Context, req *v1.InviteCodeRequest) (string, error) {
	// check username
	inviteCode := v1.GenerateNumber(5)
	user, err := s.userRepo.GetByUserId(ctx, req.UserId)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByUserId err = %s", err.Error())
		return inviteCode, v1.ErrInternalServerError
	}
	if user == nil {
		return inviteCode, v1.ErrUserNotExist
	}
	if user.Code != "" {
		return user.Code, fmt.Errorf("Already bound invitation relationship")
	}
	inviteuser, err := s.userRepo.GetByCode(ctx, req.Code)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByCode err = %s", err.Error())
		return inviteCode, v1.ErrInternalServerError
	}
	if err == nil && inviteuser == nil {
		return inviteCode, v1.ErrInvalidInviteCode
	}

	user.Code = inviteCode
	user.InviteBy = inviteuser.UserId

	// Transaction demo
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		// 创建用户
		if err = s.userRepo.Update(ctx, user); err != nil {
			return err
		}
		//插入直接邀请关系
		if err = s.inviteRepo.Create(ctx, &model.Invitations{
			InviterID: inviteuser.UserId,
			InviteeID: user.UserId,
			Level:     1,
		}); err != nil {
			return err
		}
		// 插入间接邀请关系
		invitations, err := s.inviteRepo.FindByInviteeId(ctx, inviteuser.UserId, 15)
		if err != nil {
			return err
		}
		for _, invitation := range *invitations {
			if err = s.inviteRepo.Create(ctx, &model.Invitations{
				InviterID: invitation.InviterID,
				InviteeID: user.UserId,
				Level:     invitation.Level + 1,
			}); err != nil {
				return err
			}
		}
		//插入团队纪录
		if err = s.teamRepo.Create(ctx, &model.Teamcounts{
			UserId: user.UserId,
		}); err != nil {
			return err
		}
		// 更新团队人数
		teamCounts, err := s.teamRepo.FindByInviteeId(ctx, user.UserId)
		if err != nil {
			return err
		}
		for _, teamCount := range *teamCounts {
			teamCount.TeamCount++
			if err := s.teamRepo.Update(ctx, &teamCount); err != nil {
				return err
			}
		}

		// TODO: other repo
		return nil
	})
	return inviteCode, err
}

func (s *userService) Select(ctx context.Context, user_id string) (*v1.SelectResponseData, error) {
	repo := &v1.SelectResponseData{}
	user, err := s.userRepo.GetByUserId(ctx, user_id)
	if err != nil {
		s.logger.WithContext(ctx).Sugar().Errorf("GetByUserId err = %s", err.Error())
		return repo, v1.ErrInternalServerError
	}
	if user == nil {
		return repo, v1.ErrUserNotExist
	}
	globalNodes, err := s.userRepo.GetAllNodesCount(ctx)
	if err != nil {
		return repo, err
	}
	invicount, err := s.inviteRepo.GetInviteCount(ctx, user.UserId)
	if err != nil {
		return repo, err
	}

	Amount, err := s.inviteRepo.GetInviteAmount(ctx, user.UserId)
	if err != nil {
		return repo, err
	}
	var teamreward, claimteamreward int
	if user.Address == "EFxTbjFeGtWrShQ1Gb3Z2tCujRWKsFULBaRJuycQhMMW" {
		teamreward = user.TeamReward
		claimteamreward = user.ClaimTeamReward
	}
	repo.GlobalNodes = int(globalNodes)
	repo.EvmAddress = user.EvmAddress
	repo.MyNodes = user.Nodes
	repo.EstimatedEarnings = 0
	repo.MyStable = user.Nodes
	repo.MyRacehorses = user.Nodes
	repo.DirectPerformance = (user.DirectReward + user.ClaimDirectReward) * 10
	repo.DirectReward = user.DirectReward
	repo.ClaimDirectReward = user.ClaimDirectReward
	repo.TeamPerformance = (user.TeamReward + user.ClaimTeamReward) * 20
	repo.TeamReward = teamreward
	repo.ClaimTeamReward = claimteamreward
	repo.DirectCount = invicount.DirectCount
	repo.TeamCount = invicount.TeamCount
	repo.DirectSolAmount = fmt.Sprintf("%.2f", Amount.DirectAmount)
	repo.TeamSolAmount = fmt.Sprintf("%.2f", Amount.TeamAmount)
	repo.Bonus = Amount.DirectAmount * 0.05
	return repo, nil
}
func (s *userService) Solsearch(ctx context.Context, user_id string, address string) (*[]v1.SolSearchResponData, error) {
	repos := []v1.SolSearchResponData{}
	if address != "" {
		user, err := s.userRepo.GetByAddress(ctx, address)
		if err != nil {
			return &repos, err
		}
		if user == nil {
			return &repos, nil
		}
		user_id = user.UserId
	}
	users, err := s.userRepo.SolSearch(ctx, user_id)
	if err != nil {
		return &repos, err
	}
	for _, user := range *users {
		InviteCount, err := s.inviteRepo.GetInviteCount(ctx, user.UserId)
		fmt.Printf("InviteCount:%+v\n", InviteCount)
		if err != nil {
			s.logger.WithContext(ctx).Sugar().Errorf("GetInviteCount EvmAddress:%s Err:%s", user.EvmAddress, err.Error())
			continue
		}
		repos = append(repos, v1.SolSearchResponData{
			UserId:            user.UserId,
			EvmAddress:        user.Address,
			DirectCount:       InviteCount.DirectCount,
			TeamCount:         InviteCount.TeamCount,
			DirectPerformance: (user.ClaimDirectReward + user.DirectReward) * 10,
			TeamPerformance:   (user.TeamReward + user.ClaimTeamReward) * 20,
		})
	}

	return &repos, nil
}

func (s *userService) GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error) {
	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetProfileResponseData{
		UserId: user.UserId,
	}, nil
}

func (s *userService) BindEvmAddress(ctx context.Context, userId string, req v1.BindEvmAddressRequest) error {
	if userId == "" {
		return fmt.Errorf("userId cannot be empty")
	}

	user, err := s.userRepo.GetByUserId(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.EvmAddress != "" {
		return fmt.Errorf("EvmAddress already exists")
	}

	user.EvmAddress = req.EvmAddress
	if err = s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
