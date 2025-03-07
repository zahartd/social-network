package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/zahartd/social-network/src/services/user-service/internal/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(login, firstname, surname, email, password string) (*models.User, error)
	Login(login, password, ipAddress string) (*models.User, string, error)
	GetUserByID(id string) (*models.User, error)
	GetUserByLogin(login string) (*models.User, error)
	UpdateUser(id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error)
	DeleteUser(id string, requesterID string) error
}

type userService struct {
	repo        repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewUserService(repo repository.UserRepository, sessionRepo repository.SessionRepository) UserService {
	return &userService{repo: repo, sessionRepo: sessionRepo}
}

func (s *userService) CreateUser(login, firstname, surname, email, password string) (*models.User, error) {
	existingUser, _ := s.repo.GetByLogin(login)
	if existingUser != nil {
		return nil, errors.New("user with this login already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID:           uuid.NewString(),
		Login:        login,
		Firstname:    firstname,
		Surname:      surname,
		Email:        email,
		PasswordHash: string(hash),
		Salt:         "",
	}
	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(login, password, ipAddress string) (*models.User, string, error) {
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		return nil, "", errors.New("invalid login or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", errors.New("invalid login or password")
	}
	return user, "", nil
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUserByLogin(login string) (*models.User, error) {
	return s.repo.GetByLogin(login)
}

func (s *userService) UpdateUser(id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error) {
	if id != requesterID {
		return nil, errors.New("unauthorized: cannot update another user's profile")
	}
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Email = email
	user.Firstname = firstname
	user.Surname = surname
	user.Phone = phone
	user.Bio = bio
	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) DeleteUser(id string, requesterID string) error {
	if id != requesterID {
		return errors.New("unauthorized: cannot delete another user's account")
	}
	return s.repo.Delete(id)
}
