package repository

import "github.com/zahartd/social-network/src/services/user-service/internal/domain/models"

type Session interface {
	Create(s *models.Session) error
	GetByToken(token string) (*models.Session, error)
	DeleteByToken(token string) error
}
