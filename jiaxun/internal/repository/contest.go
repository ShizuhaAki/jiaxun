package repository

import (
	"jiaxun/internal/model"

	"gorm.io/gorm"
)

type ContestRepository struct {
	*BaseRepository[model.Contest]
	db *gorm.DB
}

func NewContestRepository(db *gorm.DB) *ContestRepository {
	return &ContestRepository{
		BaseRepository: NewBaseRepository[model.Contest](db),
		db:             db,
	}
}

// Contest-specific methods
func (r *ContestRepository) GetRegistrationsByContestID(contestID uint) ([]model.ContestRegistration, error) {
	var registrations []model.ContestRegistration
	err := r.db.Where("contest_id = ?", contestID).Find(&registrations).Error
	if err != nil {
		return nil, err
	}
	return registrations, nil
}

func (r *ContestRepository) CreateRegistration(registration *model.ContestRegistration) error {
	return r.db.Create(registration).Error
}