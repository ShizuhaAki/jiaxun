package service

import (
	"errors"
	"time"

	"jiaxun/internal/model"
	"jiaxun/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// HashPassword generates a bcrypt hash from a password string
func HashPassword(password string) (string, error) {
	// Generate the hash with a cost factor (bcrypt cost of 10 is a good trade-off)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a plaintext password with a hashed one
// If this returns a nil, check passes
func CheckPassword(provided, recorded string) error {
	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(recorded), []byte(provided)); err != nil {
		return err
	}
	return nil
}

// UserService handles business logic for user operations
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Create registers a new user
func (s *UserService) Create(user *model.User) error {
	// Check if user with same username already exists
	existingUser, err := s.repo.GetByUsername(user.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	// Check if user with same email already exists
	existingEmail, err := s.repo.GetByEmail(user.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existingEmail != nil {
		return ErrEmailAlreadyExists
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now

	// Password hashing
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return s.repo.Create(user)
}

func (s *UserService) Exists(id uint) (bool, error) {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id uint) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByUsername retrieves a user by username
func (s *UserService) GetByUsername(username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email address
func (s *UserService) GetByEmail(email string) (*model.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update updates user information
func (s *UserService) Update(user *model.User) error {
	existingUser, err := s.repo.GetByID(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Check if email is being changed and is already in use by another user
	if user.Email != existingUser.Email {
		emailUser, err := s.repo.GetByEmail(user.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if emailUser != nil && emailUser.ID != user.ID {
			return ErrEmailAlreadyExists
		}
	}

	// If password changed, hash it
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	return s.repo.Update(user)
}

// Delete removes a user by ID
func (s *UserService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// Authenticate validates user credentials and returns user if valid
func (s *UserService) Authenticate(usernameOrEmail, password string) (*model.User, error) {
	var user *model.User
	var err error

	// Try to authenticate by username first
	user, err = s.repo.GetByUsername(usernameOrEmail)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// If not found by username, try by email
		user, err = s.repo.GetByEmail(usernameOrEmail)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInvalidCredentials
			}
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if err := CheckPassword(password, user.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// List returns paginated users
func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.List(page, pageSize)
}

// SearchByEmail finds users with matching email pattern
func (s *UserService) SearchByEmail(search string, page, pageSize int) ([]*model.User, int64, error) {
	return s.repo.SearchByEmail(search, page, pageSize)
}

// SearchByFullName finds users with matching full name pattern
func (s *UserService) SearchByFullName(search string, page, pageSize int) ([]*model.User, int64, error) {
	return s.repo.SearchByFullName(search, page, pageSize)
}
