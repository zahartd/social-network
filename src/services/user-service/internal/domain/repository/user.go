package repository

import "github.com/zahartd/social-network/src/services/user-service/internal/domain/models"

type User interface {
	Create(u *models.User) error
	GetByLogin(login string) (*models.User, error)
	GetByID(id string) (*models.User, error)
	Update(u *models.User) error
	Delete(id string) error
}
