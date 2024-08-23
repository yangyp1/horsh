package repository

import (
	"SolProject/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type TeamcountsRepository interface {
	FirstById(id int64) (*model.Teamcounts, error)
	Create(ctx context.Context, team *model.Teamcounts) error
	Update(ctx context.Context, teamcount *model.Teamcounts) error
	FindByInviteeId(ctx context.Context, InviteeId string) (*[]model.Teamcounts, error)
}

type teamcountsRepository struct {
	*Repository
}

func NewTeamcountsRepository(repository *Repository) TeamcountsRepository {
	return &teamcountsRepository{
		Repository: repository,
	}
}

func (r *teamcountsRepository) FirstById(id int64) (*model.Teamcounts, error) {
	var teamcounts model.Teamcounts
	// TODO: query db
	return &teamcounts, nil
}

func (r *teamcountsRepository) Create(ctx context.Context, team *model.Teamcounts) error {
	if err := r.DB(ctx).Model(model.Teamcounts{}).Create(team).Error; err != nil {
		return err
	}
	return nil
}

func (r *teamcountsRepository) Update(ctx context.Context, teamcount *model.Teamcounts) error {
	if err := r.DB(ctx).Save(teamcount).Error; err != nil {
		return err
	}
	return nil
}

func (r *teamcountsRepository) FindByInviteeId(ctx context.Context, InviteeId string) (*[]model.Teamcounts, error) {
	teamCounts := &[]model.Teamcounts{}
	if err := r.DB(ctx).Model(model.Teamcounts{}).Where("user_id IN (select inviter_id from invitations where invitee_id = ? AND level <= 15)", InviteeId).Find(teamCounts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return teamCounts, nil
		}
		return teamCounts, err
	}
	return teamCounts, nil
}
