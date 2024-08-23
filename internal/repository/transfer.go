package repository

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type TransferRepository interface {
	FirstById(id int64) (*model.Transfer, error)
	Create(ctx context.Context, transfer *model.Transfer) error
	GetByAll(ctx context.Context, req *v1.ExportRequest) (*[]v1.ExportResponse, error)
	GetByAllTeam(ctx context.Context, req *v1.ExportTeamRequest) (*[]v1.ExportResponsezTeam, error)
	GetByAllRegis(ctx context.Context) (*[]v1.ExportResponsezTeam, error)
}
type transferRepository struct {
	*Repository
}

func NewTransferRepository(repository *Repository) TransferRepository {
	return &transferRepository{
		Repository: repository,
	}
}

func (r *transferRepository) FirstById(id int64) (*model.Transfer, error) {
	var transfer model.Transfer
	// TODO: query db
	return &transfer, nil
}

func (r *transferRepository) Create(ctx context.Context, transfer *model.Transfer) error {
	if err := r.DB(ctx).Model(model.Transfer{}).Create(transfer).Error; err != nil {
		return err
	}
	return nil
}

func (r *transferRepository) GetByAll(ctx context.Context, req *v1.ExportRequest) (*[]v1.ExportResponse, error) {
	transfers := []v1.ExportResponse{}
	db := r.DB(ctx).Model(model.Transfer{}).Select("sour_addr,amount,transfer_time,users.id,users.user_id,users.code,users.invite_by")
	if req.Address != "" {
		db.Where("sour_addr = ?", req.Address)
	}
	if req.StartTime != "" {
		db.Where("transfer.transfer_time > ?", req.StartTime)
	}
	if req.EndTime != "" {
		db.Where("transfer.transfer_time < ?", req.EndTime)
	}
	db.Joins("left join users on users.address = transfer.sour_addr")
	err := db.Find(&transfers).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &transfers, nil
		}
		return &transfers, err
	}
	return &transfers, nil
}

func (r *transferRepository) GetByAllTeam(ctx context.Context, req *v1.ExportTeamRequest) (*[]v1.ExportResponsezTeam, error) {
	transfers := []v1.ExportResponsezTeam{}
	err := r.DB(ctx).Raw("select address,sol_amount,users.created_at,users.id,user_id,code,invite_by from invitations join users on users.user_id = invitations.invitee_id where inviter_id = ?", req.UserId).Scan(&transfers).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &transfers, nil
		}
		return &transfers, err
	}
	return &transfers, nil
}

func (r *transferRepository) GetByAllRegis(ctx context.Context) (*[]v1.ExportResponsezTeam, error) {
	transfers := []v1.ExportResponsezTeam{}
	err := r.DB(ctx).Raw("select address,sol_amount,created_at,users.id,user_id,code,invite_by from users").Scan(&transfers).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &transfers, nil
		}
		return &transfers, err
	}
	return &transfers, nil
}
