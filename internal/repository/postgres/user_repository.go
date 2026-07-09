package postgres

import (
	"context"
	"crm-project/internal/models/entity"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	dbGorm *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{dbGorm: db}
}

func (r *UserRepository) CreateUser(user *entity.User) (*entity.User, error) {
	isCreate := r.dbGorm.Create(user)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	ctx := context.Background()
	var result entity.User

	err := r.dbGorm.WithContext(ctx).
		Where("email = ?", email).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *UserRepository) FindByID(userID string) (*entity.User, error) {
	ctx := context.Background()
	var user entity.User
	err := r.dbGorm.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdatePassword(userID string, hashedPassword string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
	return err
}

func (r *UserRepository) Update2FASecret(userID string, secret string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("two_factor_secret", secret).Error
	return err
}

func (r *UserRepository) Enable2FA(userID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("is_two_factor_enabled", true).Error
	return err
}

func (r *UserRepository) UpdateLastLogin(userID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("last_login", time.Now()).Error
	return err
}

func (r *UserRepository) ToggleOnlineStatus(userID string, isOnline bool) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("is_online", isOnline).Error
	return err
}

func (r *UserRepository) SearchUsers(keyword string) ([]entity.User, error) {
	ctx := context.Background()
	var users []entity.User
	err := r.dbGorm.WithContext(ctx).Where("name ILIKE ?", "%"+keyword+"%").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAdmins() ([]entity.User, error) {
	ctx := context.Background()
	var users []entity.User
	err := r.dbGorm.WithContext(ctx).Where("role = ?", "admin").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAllUsers() ([]entity.User, error) {
	var users []entity.User
	result := r.dbGorm.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepository) GetFirstUser() (*entity.User, error) {
	ctx := context.Background()
	var user entity.User
	err := r.dbGorm.WithContext(ctx).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CheckIsManager(userID string) bool {
	var count int64
	r.dbGorm.Table("teams").Where("manager_id = ? AND deleted_at IS NULL", userID).Count(&count)
	return count > 0
}
