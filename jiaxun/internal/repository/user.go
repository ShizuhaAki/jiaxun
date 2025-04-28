package repository

import (
	"jiaxun/internal/model"

	"gorm.io/gorm"
)

// UserRepository provides user-specific database operations.
type UserRepository struct {
	*BaseRepository[model.User]          // Embed the BaseRepository
	db                          *gorm.DB // Keep db if we need custom queries
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[model.User](db),
		db:             db,
	}
}

// --- User-specific methods ---

// GetByUsername retrieves a user by their username.
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SearchByEmail finds users whose email contains the search string.
func (r *UserRepository) SearchByEmail(search string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := r.db.Model(&model.User{}).Where("email LIKE ?", "%"+search+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SearchByFullName finds users whose full_name contains the search string.
func (r *UserRepository) SearchByFullName(search string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := r.db.Model(&model.User{}).Where("full_name LIKE ?", "%"+search+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
