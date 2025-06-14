package service

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/domain/repository"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/auth"
)

//go:generate mockgen -source=user.go -destination=../../transport/http/mocks/event_publisher_mock.go -package=mocks

type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, ev models.User) error
}

type Service struct {
	users     repository.User
	sessions  repository.Session
	publisher EventPublisher
}

func NewUser(u repository.User, s repository.Session, p EventPublisher) *Service {
	return &Service{u, s, p}
}

func (s *Service) Create(login, firstname, surname, email, password string) (*models.User, string, error) {
	if existing, _ := s.users.GetByLogin(login); existing != nil {
		return nil, "", errors.New("user already exists")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.NewString(),
		Login:        login,
		Firstname:    firstname,
		Surname:      surname,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(user); err != nil {
		return nil, "", err
	}
	token, _ := auth.GenerateToken(user)

	_ = s.publisher.PublishUserRegistered(context.TODO(), *user) // fire‑and‑forget

	return user, token, nil
}

func (s *Service) Login(ip, login, password string) (string, error) {
	user, err := s.users.GetByLogin(login)
	if err != nil {
		return "", errors.New("invalid login or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid login or password")
	}
	token, _ := auth.GenerateToken(user)

	sess := &models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(auth.TokenTTL),
		IPAddress: net.ParseIP(ip),
	}
	_ = s.sessions.Create(sess)
	return token, nil
}

func (s *Service) Logout(token string) error {
	return s.sessions.DeleteByToken(token)
}

func (s *Service) GetByID(id string) (*models.User, error)   { return s.users.GetByID(id) }
func (s *Service) GetByLogin(l string) (*models.User, error) { return s.users.GetByLogin(l) }

func (s *Service) Update(id string, req *models.User) (*models.User, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Email = req.Email
	user.Firstname = req.Firstname
	user.Surname = req.Surname
	user.Phone = req.Phone
	user.Bio = req.Bio
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Delete(id, token string) error {
	if err := s.sessions.DeleteByToken(token); err != nil {
		return err
	}
	return s.users.Delete(id)
}
