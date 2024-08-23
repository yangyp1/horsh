package repository

import (
	v1 "SolProject/api/v1"
	"SolProject/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	UpDateByHorshCount(ctx context.Context, userId string) error
	UpDateByUSDTCount(ctx context.Context, userId string) error
	UpDateByTeamReward(ctx context.Context, userId string) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByAddress(ctx context.Context, address string) (*model.User, error)
	GetByEvmAddress(ctx context.Context, address string) (*model.User, error)
	GetByCode(ctx context.Context, code string) (*model.User, error)
	GetByUserId(ctx context.Context, user_id string) (*model.User, error)
	SolSearch(ctx context.Context, user_id string) (*[]model.User, error)
	UpdateColumn(ctx context.Context, address string, amount float64) error
	GetByAll(ctx context.Context, req *v1.AdminSearchRequest, limit int, offset int) (int64, *[]model.User, error)
	GetAllCounty(ctx context.Context) (*v1.AdminAllCountResponseData, error)
	GetAllNodesCount(ctx context.Context) (int64, error)
}

func NewUserRepository(
	r *Repository,
) UserRepository {
	return &userRepository{
		Repository: r,
	}
}

type userRepository struct {
	*Repository
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Model(model.User{}).Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpDateByHorshCount(ctx context.Context, userId string) error {
	if err := r.DB(ctx).Model(model.User{}).Where("user_id = ?", userId).UpdateColumn("horsh_count", gorm.Expr("horsh_count + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpDateByUSDTCount(ctx context.Context, userId string) error {
	if err := r.DB(ctx).Model(model.User{}).Where("user_id = ?", userId).UpdateColumn("usdt_count", gorm.Expr("usdt_count + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpDateByTeamReward(ctx context.Context, userId string) error {
	if err := r.DB(ctx).Model(model.User{}).Where("user_id = ?", userId).UpdateColumn("team_reward", gorm.Expr("team_reward + ?", 25)).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateColumn(ctx context.Context, address string, amount float64) error {
	if err := r.DB(ctx).Model(model.User{}).Where("address = ?", address).UpdateColumn("sol_amount", gorm.Expr("sol_amount + ?", amount)).Error; err != nil {
		return err
	}
	return nil
}


func (r *userRepository) GetAllNodesCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.DB(ctx).Model(model.User{}).Select("sum(nodes) as count").Take(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepository) GetByID(ctx context.Context, userId string) (*model.User, error) {
	user := model.User{}
	if err := r.DB(ctx).Model(model.User{}).Where("user_id = ?", userId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByAddress(ctx context.Context, address string) (*model.User, error) {
	user := model.User{}
	if err := r.DB(ctx).Model(model.User{}).Where("address = ?", address).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEvmAddress(ctx context.Context, address string) (*model.User, error) {
	user := model.User{}
	if err := r.DB(ctx).Model(model.User{}).Where("evm_address = ?", address).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByAll(ctx context.Context, req *v1.AdminSearchRequest, limit int, offset int) (int64, *[]model.User, error) {
	users := []model.User{}
	var count int64
	db := r.DB(ctx).Model(model.User{})
	if req.Address != "" {
		db.Where("address = ?", req.Address)
	}
	if req.Code != "" {
		db.Where("code = ?", req.Code)
	}
	if req.InviteBy != "" {
		db.Where("invite_by = ?", req.InviteBy)
	}
	if req.StartTime != "" {
		db.Where("created_at > ?", req.StartTime)
	}
	if req.EndTime != "" {
		db.Where("created_at < ?", req.EndTime)
	}
	if req.IsInviteby {
		db.Where("invite_by = ?", "")
	}
	if req.IsNotEmptyAmount {
		db.Where("sol_amount != 0")
	}
	err := db.Count(&count).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return count, nil, nil
		}
		return count, nil, err
	}
	return count, &users, nil
}

//SELECT COUNT(*) AS regis_count,SUM(CASE WHEN sol_amount != 0 THEN 1 ELSE 0 END) AS bonus_count,SUM(sol_amount) AS totol_amount FROM users;

func (r *userRepository) GetAllCounty(ctx context.Context) (*v1.AdminAllCountResponseData, error) {
	allcount := v1.AdminAllCountResponseData{}
	if err := r.DB(ctx).Raw("SELECT COUNT(*) AS regis_count,SUM(CASE WHEN sol_amount != 0 THEN 1 ELSE 0 END) AS bonus_count,SUM(sol_amount) AS totol_amount FROM users").Scan(&allcount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &allcount, nil
		}
		return &allcount, err
	}
	return &allcount, nil
}

func (r *userRepository) GetByUserId(ctx context.Context, user_id string) (*model.User, error) {
	user := model.User{}
	if err := r.DB(ctx).Model(model.User{}).Where("user_id = ?", user_id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByCode(ctx context.Context, code string) (*model.User, error) {
	user := model.User{}
	if err := r.DB(ctx).Model(model.User{}).Where("code = ?", code).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SolSearch(ctx context.Context, user_id string) (*[]model.User, error) {
	user := []model.User{}
	if err := r.DB(ctx).Model(model.User{}).Select("users.*").Joins("join invitations on users.user_id = invitee_id").Where("invitations.inviter_id = ? and level = 1", user_id).Find(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil

}
