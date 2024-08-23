package handler

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/service"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

type UserHandler struct {
	*Handler
	userService service.UserService
}

func NewUserHandler(handler *Handler, userService service.UserService) *UserHandler {
	return &UserHandler{
		Handler:     handler,
		userService: userService,
	}
}

// Login godoc
// @Summary 账号登录
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body v1.LoginRequest true "params"
// @Success 200 {object} v1.LoginResponse
// @Router /login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req v1.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	user, token, err := h.userService.Login(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, v1.LoginResponseData{
		AccessToken: token,
		Address:     user.Address,
		UserId:      user.UserId,
		Code:        user.Code,
		SolAmount:   user.SolAmount,
		EvmAddress:  user.EvmAddress,
		HorshCount:  user.HorshCount,
		UsdtCount:   user.UsdtCount,
	})
}

// Login godoc
// @Summary 使用邀请码生成邀请链接
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.InviteCodeRequest true "params"
// @Success 200 {object} v1.InviteCodeResponse
// @Router /user/Invitecode [post]
func (h *UserHandler) Invitecode(ctx *gin.Context) {
	var req v1.InviteCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	code, err := h.userService.InviteCode(ctx, &req)
	if err != nil {
		h.logger.WithContext(ctx).Sugar().Errorf("生成邀请链接失败%s", err.Error())
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, v1.InviteCodeResponseData{
		UserId: req.UserId,
		Code:   code,
	})
}

// Login godoc
// @Summary 邀请人数查询
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id query string true "用户ID"
// @Success 200 {object} v1.SelectResponse
// @Router /user/select [get]
func (h *UserHandler) Select(ctx *gin.Context) {
	user_id := GetUserIdFromCtx(ctx)
	if user_id == "" {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrUnauthorized, nil)
		return
	}

	repo, err := h.userService.Select(ctx, user_id)
	if err != nil {
		h.logger.WithContext(ctx).Sugar().Errorf("查询邀请人数失败：%s", err.Error())
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, *repo)
}

// Login godoc
// @Summary 搜索SOL地址
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.SolSearchRequest true "params"
// @Success 200 {object} v1.SolSearchResponse
// @Router /user/Solsearch [post]
func (h *UserHandler) Solsearch(ctx *gin.Context) {
	user_id := GetUserIdFromCtx(ctx)
	if user_id == "" {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrUnauthorized, nil)
		return
	}
	var req v1.SolSearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	repos, err := h.userService.Solsearch(ctx, user_id, req.Address)
	if err != nil {
		h.logger.WithContext(ctx).Sugar().Errorf("Solsearch Err:%s", err.Error())
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, *repos)
}

// GetProfile godoc
// @Summary 获取用户信息
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetProfileResponse
// @Router /user [get]
func (h *UserHandler) GetProfile(ctx *gin.Context) {
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrUnauthorized, nil)
		return
	}

	user, err := h.userService.GetProfile(ctx, userId)
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	v1.HandleSuccess(ctx, user)
}

// UpdateProfile godoc
// @Summary 绑定evm地址
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.BindEvmAddressRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/bind/evmaddress [post]
func (h *UserHandler) BindEvmAddress(ctx *gin.Context) {
	userId := GetUserIdFromCtx(ctx)
	var req v1.BindEvmAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	if err := h.userService.BindEvmAddress(ctx, userId, req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// UpdateProfile godoc
// @Summary Horsh交易确认
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.HorshTransferRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/horsh/transfer [post]
func (h *UserHandler) HorshTransfer(ctx *gin.Context) {
	userId := GetUserIdFromCtx(ctx)
	var req v1.HorshTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	user, confirm_status, err := h.userService.HorshTransfer(ctx, userId, req)
	if err != nil {
		h.logger.Sugar().Errorf("HorshTransfer Err:%s", err.Error())
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, gin.H{
		"horsh_count":    user.HorshCount+1,
		"confirm_status": confirm_status,
	})
}

// UpdateProfile godoc
// @Summary USDT交易确认
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.USDTTransferRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/usdt/transfer [post]
func (h *UserHandler) USDTTransfer(ctx *gin.Context) {
	userId := GetUserIdFromCtx(ctx)
	var req v1.USDTTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	user, confirm_status, err := h.userService.USDTTransfer(ctx, userId, req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, gin.H{
		"usdt_count":     user.UsdtCount+1,
		"confirm_status": confirm_status,
	})
}

// ClaimReward godoc
// @Summary 提取奖励
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ClaimRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/claim/reward [post]
func (h *UserHandler) ClaimReward(ctx *gin.Context) {
	userId := GetUserIdFromCtx(ctx)
	var req v1.ClaimRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	user, err := h.userService.ClaimReward(ctx, userId, req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, gin.H{
		"direct_reward":       user.DirectReward,
		"claim_direct_reward": user.ClaimDirectReward,
		"team_reward":         user.TeamReward,
		"claim_team_reward":   user.ClaimTeamReward,
	})
}

// Login godoc
// @Summary 登录
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param request body v1.AdminLoginRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/login [post]
func (h *UserHandler) AdminLogin(ctx *gin.Context) {
	var req v1.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if req.Username != "horsh" || req.Password != "horsh" {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrNotFound, nil)
		return
	}
	token, err := h.userService.AdminLogin(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, gin.H{"accessToken": token})
}

// Login godoc
// @Summary 查询
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.AdminSearchRequest true "params"
// @Success 200 {object} v1.AdminSearchResponse
// @Router /admin/search [post]
func (h *UserHandler) AdminSearch(ctx *gin.Context) {
	var req v1.AdminSearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 30
	}
	offset := (req.Page - 1) * req.PageSize
	count, repo, err := h.userService.AdminSearch(ctx, &req, req.PageSize, offset)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, v1.AdminSearchResponseDa{
		Count:      count,
		ReturnData: *repo,
	})
}

// Login godoc
// @Summary 导出转账记录
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ExportRequest true "params"
// @Router /admin/export-record [post]
func (h *UserHandler) ExportRecord(ctx *gin.Context) {
	var req v1.ExportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	repo, err := h.userService.ExportRecord(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	csvFile, err := os.Create("TransferRecord.csv")
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	defer csvFile.Close()

	// 写入CSV数据
	if err := gocsv.MarshalFile(repo, csvFile); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename=TransferRecord.csv")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File("TransferRecord.csv")

}

// Login godoc
// @Summary 导出转账记录整个团队
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ExportTeamRequest true "params"
// @Router /admin/export-record-team [post]
func (h *UserHandler) ExportRecordTeam(ctx *gin.Context) {
	var req v1.ExportTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	repo, err := h.userService.ExportRecordTeam(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	csvFile, err := os.Create("TransferRecord.csv")
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	defer csvFile.Close()
	// 写入CSV数据
	if err := gocsv.MarshalFile(repo, csvFile); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename=TransferRecord.csv")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File("TransferRecord.csv")
}

// Login godoc
// @Summary 导出所有注册人的记录
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Router /admin/export-record-regis [post]
func (h *UserHandler) ExportRecordRegis(ctx *gin.Context) {
	repo, err := h.userService.ExportRecordRegis(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	csvFile, err := os.Create("TransferRecord.csv")
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	defer csvFile.Close()
	// 写入CSV数据
	if err := gocsv.MarshalFile(repo, csvFile); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename=TransferRecord.csv")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File("TransferRecord.csv")
}

// Login godoc
// @Summary 查询总注册人数
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.AllCountResponse
// @Router /admin/allCount [get]
func (h *UserHandler) AdminAllCount(ctx *gin.Context) {
	repo, err := h.userService.AdminAllCount(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, *repo)
}

// Login godoc
// @Summary 查询邀请码邀请总数
// @Schemes
// @Description
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param code query string true "分页"
// @Security Bearer
// @Router /admin/CreationCount [get]
func (h *UserHandler) CreationCount(ctx *gin.Context) {
	if ctx.Query("code") == "" {
		err := errors.New("code is Empty")
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	count, err := h.userService.CreationCount(ctx, ctx.Query("code"))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, gin.H{
		"count": count,
	})
}
