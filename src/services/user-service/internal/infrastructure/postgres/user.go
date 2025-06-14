package postgres

import (
	"database/sql"
	"errors"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) Create(u *models.User) error {
	q := `INSERT INTO users (id,login,firstname,surname,email,phone,bio,password_hash,created_at,updated_at)
          VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.Exec(q, u.ID, u.Login, u.Firstname, u.Surname, u.Email, u.Phone, u.Bio, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepo) GetByLogin(l string) (*models.User, error) {
	row := r.db.QueryRow(`SELECT id,login,firstname,surname,email,phone,bio,password_hash,created_at,updated_at FROM users WHERE login=$1`, l)
	var u models.User
	err := row.Scan(&u.ID, &u.Login, &u.Firstname, &u.Surname, &u.Email, &u.Phone, &u.Bio, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return &u, err
}

func (r *UserRepo) GetByID(id string) (*models.User, error) {
	row := r.db.QueryRow(`SELECT id,login,firstname,surname,email,phone,bio,password_hash,created_at,updated_at FROM users WHERE id=$1`, id)
	var u models.User
	err := row.Scan(&u.ID, &u.Login, &u.Firstname, &u.Surname, &u.Email, &u.Phone, &u.Bio, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return &u, err
}

func (r *UserRepo) Update(u *models.User) error {
	res, err := r.db.Exec(`UPDATE users SET email=$1,firstname=$2,surname=$3,phone=$4,bio=$5,updated_at=$6 WHERE id=$7`,
		u.Email, u.Firstname, u.Surname, u.Phone, u.Bio, u.UpdatedAt, u.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("user not found")
	}
	return nil
}
