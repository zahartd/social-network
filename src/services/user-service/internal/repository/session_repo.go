package repository

import (
	"database/sql"
	"errors"
	"net"

	"github.com/zahartd/social-network/src/services/user-service/internal/models"
)

type SessionRepository interface {
	CreateSession(session *models.Session) error
	GetSessionByToken(token string) (*models.Session, error)
	DeleteSessionByToken(token string) error
}

type postgresSessionRepo struct {
	db *sql.DB
}

func NewPostgresSessionRepo(db *sql.DB) SessionRepository {
	return &postgresSessionRepo{db: db}
}

func (r *postgresSessionRepo) CreateSession(session *models.Session) error {
	query := `
	INSERT INTO user_sessions (user_id, token, created_at, expires_at, ip_address)
	VALUES ($1, $2, now(), $3, $4)
	RETURNING id, created_at`
	return r.db.QueryRow(query, session.UserID, session.Token, session.ExpiresAt, session.IPAddress).
		Scan(&session.ID, &session.CreatedAt)
}

func (r *postgresSessionRepo) GetSessionByToken(token string) (*models.Session, error) {
	query := `
	SELECT id, user_id, token, created_at, expires_at, ip_address
	FROM user_sessions WHERE token=$1`
	row := r.db.QueryRow(query, token)
	session := &models.Session{}
	var ipStr string
	err := row.Scan(&session.ID, &session.UserID, &session.Token, &session.CreatedAt, &session.ExpiresAt, &ipStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	session.IPAddress = net.ParseIP(ipStr)
	return session, nil
}

func (r *postgresSessionRepo) DeleteSessionByToken(token string) error {
	query := `DELETE FROM user_sessions WHERE token=$1`
	result, err := r.db.Exec(query, token)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("session not found")
	}
	return nil
}
