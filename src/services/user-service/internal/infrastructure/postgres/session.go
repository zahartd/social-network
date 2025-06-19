package postgres

import (
	"database/sql"
	"errors"
	"net"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

type SessionRepo struct{ db *sql.DB }

func NewSessionRepo(db *sql.DB) *SessionRepo { return &SessionRepo{db} }

func (r *SessionRepo) Create(s *models.Session) error {
	q := `INSERT INTO user_sessions (user_id,token,created_at,expires_at,ip_address)
          VALUES ($1,$2,now(),$3,$4) RETURNING id,created_at`
	ip := s.IPAddress.String()
	return r.db.QueryRow(q, s.UserID, s.Token, s.ExpiresAt, ip).Scan(&s.ID, &s.CreatedAt)
}

func (r *SessionRepo) GetByToken(tok string) (*models.Session, error) {
	row := r.db.QueryRow(`SELECT id,user_id,token,created_at,expires_at,ip_address FROM user_sessions WHERE token=$1`, tok)
	var s models.Session
	var ip string
	err := row.Scan(&s.ID, &s.UserID, &s.Token, &s.CreatedAt, &s.ExpiresAt, &ip)
	if err == sql.ErrNoRows {
		return nil, errors.New("session not found")
	}
	s.IPAddress = net.ParseIP(ip)
	return &s, err
}

func (r *SessionRepo) DeleteByToken(tok string) error {
	res, err := r.db.Exec(`DELETE FROM user_sessions WHERE token=$1`, tok)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("session not found")
	}
	return nil
}
