package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	CreateUser(ctx *gin.Context, login, firstname, surname, email, password string) (*models.User, string, error)
	Login(ctx *gin.Context, login, password string) (string, error)
	Logout(ctx *gin.Context, token string) error
	GetUserByID(ctx *gin.Context, id string) (*models.User, error)
	GetUserByLogin(ctx *gin.Context, login string) (*models.User, error)
	UpdateUser(ctx *gin.Context, id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error)
	DeleteUser(ctx *gin.Context, id, token string) error
}

type userService struct {
	repo                repository.UserRepository
	sessionRepo         repository.SessionRepository
	registrationsWriter *kafka.Writer
}

func NewUserService(repo repository.UserRepository, sessionRepo repository.SessionRepository, rw *kafka.Writer) UserService {
	return &userService{
		repo:                repo,
		sessionRepo:         sessionRepo,
		registrationsWriter: rw,
	}
}

func (s *userService) CreateUser(ctx *gin.Context, login, firstname, surname, email, password string) (*models.User, string, error) {
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
		return nil, "", errors.New("failed to generate auth token")
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
	payload, _ := json.Marshal(ev)

	const retries = 3
	for range retries {
		writerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err = s.registrationsWriter.WriteMessages(
			writerCtx,
			kafka.Message{
				Key:   []byte(user.ID),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Printf("failed to write messages: %s", err.Error())
		}
		break
	}
	return user, token, nil
}

func (s *userService) Login(ctx *gin.Context, login, password string) (string, error) {
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		return "", errors.New("invalid login")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}
	token, err := auth.GenerateToken(user)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	if err := auth.CreateSession(user, token, ctx.ClientIP()); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return token, nil
}

func (s *userService) Logout(ctx *gin.Context, token string) error {
	token = auth.TrimBearerPrefix(token)
	if err := auth.DeleteSession(token); err != nil {
		return errors.New("failed to delete session")
	}
	return nil
}

func (s *userService) GetUserByID(ctx *gin.Context, id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUserByLogin(ctx *gin.Context, login string) (*models.User, error) {
	return s.repo.GetByLogin(login)
}

func (s *userService) UpdateUser(ctx *gin.Context, id string, email, firstname, surname, phone, bio string, requesterID string) (*models.User, error) {
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

func (s *userService) DeleteUser(ctx *gin.Context, id, token string) error {
	if err := s.sessionRepo.DeleteSessionByToken(token); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
