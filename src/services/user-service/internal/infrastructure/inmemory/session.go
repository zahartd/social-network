package inmemory

import (
	"errors"
	"sync"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

type InMemorySessionRepo struct {
	mu       sync.Mutex
	sessions map[string]*models.Session // key: token
}

func NewInMemorySessionRepo() *InMemorySessionRepo {
	return &InMemorySessionRepo{
		sessions: make(map[string]*models.Session),
	}
}

func (r *InMemorySessionRepo) Create(sess *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copySess := *sess
	r.sessions[sess.Token] = &copySess
	return nil
}

func (r *InMemorySessionRepo) GetByToken(token string) (*models.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, ok := r.sessions[token]
	if !ok {
		return nil, errors.New("session not found")
	}
	copySess := *sess
	return &copySess, nil
}

func (r *InMemorySessionRepo) DeleteByToken(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[token]; !ok {
		return errors.New("session not found")
	}
	delete(r.sessions, token)
	return nil
}
