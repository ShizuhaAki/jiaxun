package service

import (
	"errors"
	"jiaxun/internal/model"
	"jiaxun/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrContestAlreadyExists = errors.New("contest already exists")
	ErrContestNotFound      = errors.New("contest not found")
	ErrNotImplemented       = errors.New("not implemented")
)

type ContestService struct {
	repo *repository.ContestRepository
	userService *UserService
}

func NewContestService(repo *repository.ContestRepository) *ContestService {
	return &ContestService{
		repo: repo,
	}
}

func (s *ContestService) CreateContest(contest *model.Contest) error {
	err := s.repo.Create(contest)
	if err != nil { // wrap around gorm errors
		// duplicate entry error handling
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrContestAlreadyExists
		}
		return err
	}
	return nil
}

func (s *ContestService) GetContestByID(id uint) (*model.Contest, error) {
	contest, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestNotFound
		}
		return nil, err
	}
	return contest, nil
}

func (s *ContestService) ListContests(page, pageSize int) ([]model.Contest, int64, error) {
	contests, total, err := s.repo.List(page, pageSize)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrContestNotFound
		}
		return nil, 0, err
	}
	return contests, total,  nil
}


func (s *ContestService) UpdateContest(contest *model.Contest) error {
	existingContest, err := s.repo.GetByID(contest.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContestNotFound
		}
		return err
	}

	return s.repo.Update(existingContest)
}


// Register a user to a contest
func (s *ContestService) RegisterUserToContest(contestID uint, userID uint) error {
	// Check if the contest exists
	contest, err := s.repo.GetByID(contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContestNotFound
		}
		return err
	}

	// Check if the user exists
	exists, err := s.userService.Exists(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	// Create a new registration
	registration := &model.ContestRegistration{
		UserID:    &userID,
		ContestID: contest.ContestID,
	}

	return s.repo.CreateRegistration(registration)
}


func (s *ContestService) RegisterTeamToContest(contestID, teamID uint) error {
	// Currently not supported
	return ErrNotImplemented
	// Check if the contest exists
	// contest, err := s.repo.GetByID(contestID)
	// if err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		return ErrContestNotFound
	// 	}
	// 	return err
	// }

	// // Create a new registration
	// registration := &model.ContestRegistration{
	// 	TeamID: &teamID,
	// 	ContestID: contest.ContestID,
	// }
	// return s.repo.CreateRegistration(registration)
}

