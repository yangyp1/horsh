package handler

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)



type TeamcountsHandler struct {
	*Handler
	teamcountsService service.TeamcountsService
}

func NewTeamcountsHandler(handler *Handler, teamcountsService service.TeamcountsService) *TeamcountsHandler {
	return &TeamcountsHandler{
		Handler:     handler,
		teamcountsService: teamcountsService,
	}
}

func (h *TeamcountsHandler) GetTeamcountsById(ctx *gin.Context) {
	var params struct {
		Id int64 `form:"id" binding:"required"`
	}
	if err := ctx.ShouldBind(&params); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, err, nil)
		return
	}

	teamcounts, err := h.teamcountsService.GetTeamcountsById(params.Id)
	h.logger.Info("GetTeamcountsByID", zap.Any("teamcounts", teamcounts))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError,  err, nil)
		return
	}
	v1.HandleSuccess(ctx, teamcounts)
}

func (h *TeamcountsHandler) UpdateTeamcounts(ctx *gin.Context) {
	v1.HandleSuccess(ctx, nil)
}
