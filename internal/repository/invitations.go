package repository

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type InvitationsRepository interface {
	FirstById(id int64) (*model.Invitations, error)
	Create(ctx context.Context, invite *model.Invitations) error
	FindByInviteeId(ctx context.Context, InviteeId string, level int) (*[]model.Invitations, error)
	FindByInviterId(ctx context.Context, InviterId string, level int) (*[]model.Invitations, error)
	GetInviteCount(ctx context.Context, user_id string) (*v1.Invitecount, error)
	GetInviteAmount(ctx context.Context, user_id string) (*v1.InviteAmount, error)
	CreationCount(ctx context.Context, codes []string) ([]string, error)
}
type invitationsRepository struct {
	*Repository
}

func NewInvitationsRepository(repository *Repository) InvitationsRepository {
	return &invitationsRepository{
		Repository: repository,
	}
}

func (r *invitationsRepository) FirstById(id int64) (*model.Invitations, error) {
	var invitations model.Invitations
	// TODO: query db
	return &invitations, nil
}

func (r *invitationsRepository) Create(ctx context.Context, invite *model.Invitations) error {
	if err := r.DB(ctx).Model(model.Invitations{}).Create(invite).Error; err != nil {
		return err
	}
	return nil
}

func (r *invitationsRepository) FindByInviteeId(ctx context.Context, InviteeId string, level int) (*[]model.Invitations, error) {
	invitations := &[]model.Invitations{}
	if err := r.DB(ctx).Model(model.Invitations{}).Where("invitee_id = ? AND level < ?", InviteeId, level).Find(invitations).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return invitations, nil
		}
		return invitations, err
	}
	return invitations, nil
}

func (r *invitationsRepository) FindByInviterId(ctx context.Context, InviterId string, level int) (*[]model.Invitations, error) {
	invitations := &[]model.Invitations{}
	if err := r.DB(ctx).Model(model.Invitations{}).Where("inviter_id = ? AND level = ?", InviterId, level).Find(invitations).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return invitations, nil
		}
		return invitations, err
	}
	return invitations, nil
}

//SELECT COUNT(*) AS teamcount,COALESCE(SUM(CASE WHEN level = 1 THEN 1 ELSE 0 END),0) AS directcount FROM invitations group by inviter_id;

func (r *invitationsRepository) GetInviteCount(ctx context.Context, user_id string) (*v1.Invitecount, error) {
	invcot := v1.Invitecount{}
	if err := r.DB(ctx).Raw("SELECT COUNT(*) AS teamcount,COALESCE(SUM(CASE WHEN level = 1 THEN 1 ELSE 0 END),0) AS directcount FROM invitations WHERE inviter_id = ?", user_id).Scan(&invcot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &invcot, nil
		}
		return &invcot, err
	}
	return &invcot, nil
}

//SELECT COUNT(*) AS indirectCount,COALESCE(SUM(CASE WHEN level = 1 THEN 1 ELSE 0 END),0) AS teamCount,
// COALESCE(SUM(sol_amount),0) as indirectAmount
// FROM invitations
// join users on users.user_id = invitations.inviter_id  group by inviter_id ;

func (r *invitationsRepository) GetInviteAmount(ctx context.Context, user_id string) (*v1.InviteAmount, error) {
	amount := v1.InviteAmount{}
	if err := r.DB(ctx).Raw("SELECT COALESCE(SUM(CASE WHEN level = 1 THEN sol_amount ELSE 0 END),0) as directamount,COALESCE(SUM(sol_amount),0) as teamamount FROM invitations join users on users.user_id = invitations.invitee_id WHERE inviter_id = ?", user_id).Scan(&amount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &amount, nil
		}
		return &amount, err
	}
	return &amount, nil
}

func (r *invitationsRepository) CreationCount(ctx context.Context, codes []string) ([]string, error) {
	inviteeids := []string{}
	if err := r.DB(ctx).Raw("SELECT invitee_id FROM invitations WHERE inviter_id IN (?) and level = 1", codes).Scan(&inviteeids).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return inviteeids, nil
		}
		return inviteeids, err
	}
	return inviteeids, nil
}
