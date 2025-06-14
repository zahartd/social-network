package inmemory

import (
	"errors"
	"sync"
	"time"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

type InMemoryUserRepo struct {
	mu           sync.Mutex
	usersByID    map[string]*models.User
	usersByLogin map[string]*models.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		usersByID:    make(map[string]*models.User),
		usersByLogin: make(map[string]*models.User),
	}
}

func (r *InMemoryUserRepo) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByLogin[user.Login]; exists {
		return errors.New("user already exists")
	}

	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	uCopy := *user
	r.usersByID[user.ID] = &uCopy
	r.usersByLogin[uCopy.Login] = &uCopy
	return nil
}

func (r *InMemoryUserRepo) GetByLogin(login string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.usersByLogin[login]
	if !ok {
		return nil, errors.New("user not found")
	}
	uCopy := *u
	return &uCopy, nil
}

func (r *InMemoryUserRepo) GetByID(id string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.usersByID[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	uCopy := *u
	return &uCopy, nil
}

func (r *InMemoryUserRepo) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.usersByID[user.ID]
	if !ok {
		return errors.New("user not found")
	}
	existing.Email = user.Email
	existing.Firstname = user.Firstname
	existing.Surname = user.Surname
	existing.Phone = user.Phone
	existing.Bio = user.Bio
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *InMemoryUserRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.usersByID[id]
	if !ok {
		return errors.New("user not found")
	}
	delete(r.usersByLogin, u.Login)
	delete(r.usersByID, id)
	return nil
}
