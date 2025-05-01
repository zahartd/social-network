package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/repository"
)

type UserService interface {
	CreateUser(c *gin.Context, login, firstname, surname, email, password string) (*models.User, string, error)
	Login(c *gin.Context, login, password string) (*models.User, string, error)
	GetUserByID(c *gin.Context, id string) (*models.User, error)
	GetUserByLogin(c *gin.Context, login string) (*models.User, error)
	UpdateUser(c *gin.Context, id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error)
	DeleteUser(c *gin.Context, id, token string) error
}

type userService struct {
	repo        repository.UserRepository
	sessionRepo repository.SessionRepository
	kafkaWriter *kafka.Writer
}

func NewUserService(repo repository.UserRepository, sessionRepo repository.SessionRepository) UserService {
	return &userService{
		repo:        repo,
		sessionRepo: sessionRepo,
		kafkaWriter: kafka.NewWriter(kafka.WriterConfig{
			Brokers: []string{os.Getenv("KAFKA_BROKER_URL")},
			Topic:   "user-registrations",
			Async:   true,
		}),
	}
}

func (s *userService) CreateUser(c *gin.Context, login, firstname, surname, email, password string) (*models.User, string, error) {
	existingUser, _ := s.repo.GetByLogin(login)
	if existingUser != nil {
		return nil, "", errors.New("user with this login already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		ID:           uuid.NewString(),
		Login:        login,
		Firstname:    firstname,
		Surname:      surname,
		Email:        email,
		PasswordHash: string(hash),
	}
	err = s.repo.Create(user)
	if err != nil {
		return nil, "", err
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	ev := struct {
		UserID    string    `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		Email     string    `json:"email"`
	}{
		UserID:    user.ID,
		CreatedAt: user.CreatedAt,
		Email:     user.Email,
	}
	buf, _ := json.Marshal(ev)
	msg := kafka.Message{Key: []byte(user.ID), Value: buf}
	err = s.kafkaWriter.WriteMessages(c, msg)
	if err != nil {
		log.Printf("failed to produce event to Kafka: %s", err.Error())
	}

	return user, token, nil
}

func (s *userService) Login(c *gin.Context, login, password string) (*models.User, string, error) {
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		return nil, "", errors.New("invalid login")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", errors.New("invalid password")
	}
	token, err := auth.GenerateToken(user)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}
	if err := auth.CreateSession(user, token, c.ClientIP()); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}
	return user, token, nil
}

func (s *userService) GetUserByID(_ *gin.Context, id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUserByLogin(_ *gin.Context, login string) (*models.User, error) {
	return s.repo.GetByLogin(login)
}

func (s *userService) UpdateUser(_ *gin.Context, id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error) {
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

func (s *userService) DeleteUser(_ *gin.Context, id, token string) error {
	if err := s.sessionRepo.DeleteSessionByToken(token); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
