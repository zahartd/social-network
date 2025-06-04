package inmemory

import (
	"errors"
	"sync"

	"github.com/zahartd/social-network/src/services/user-service/internal/models"
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

func (r *InMemorySessionRepo) CreateSession(sess *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Просто кладём в map по ключу токена:
	copySess := *sess
	r.sessions[sess.Token] = &copySess
	return nil
}

func (r *InMemorySessionRepo) GetSessionByToken(token string) (*models.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, ok := r.sessions[token]
	if !ok {
		return nil, errors.New("session not found")
	}
	copySess := *sess
	return &copySess, nil
}

func (r *InMemorySessionRepo) DeleteSessionByToken(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[token]; !ok {
		return errors.New("session not found")
	}
	delete(r.sessions, token)
	return nil
}
