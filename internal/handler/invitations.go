package handler

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)


type InvitationsHandler struct {
	*Handler
	invitationsService service.InvitationsService
}

func NewInvitationsHandler(handler *Handler, invitationsService service.InvitationsService) *InvitationsHandler {
	return &InvitationsHandler{
		Handler:     handler,
		invitationsService: invitationsService,
	}
}

func (h *InvitationsHandler) GetInvitationsById(ctx *gin.Context) {
	var params struct {
		Id int64 `form:"id" binding:"required"`
	}
	if err := ctx.ShouldBind(&params); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest,  err, nil)
		return
	}

	invitations, err := h.invitationsService.GetInvitationsById(params.Id)
	h.logger.Info("GetInvitationsByID", zap.Any("invitations", invitations))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError,  err, nil)
		return
	}
	v1.HandleSuccess(ctx, invitations)
}

func (h *InvitationsHandler) UpdateInvitations(ctx *gin.Context) {
	v1.HandleSuccess(ctx, nil)
}
