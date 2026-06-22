package postgres

import (
	"context"
	"crm-project/internal/models/entity"

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
