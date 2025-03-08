package repository

import (
	"database/sql"
	"errors"

	"github.com/zahartd/social-network/src/services/user-service/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByLogin(login string) (*models.User, error)
	GetByID(id string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}

type postgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) UserRepository {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) Create(user *models.User) error {
	query := `
	INSERT INTO
	users (login, firstname, surname, email, phone, bio, password_hash, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
	RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, user.Login, user.Firstname, user.Surname, user.Email, user.Phone, user.Bio, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *postgresUserRepo) GetByLogin(login string) (*models.User, error) {
	query := `
	SELECT id, login, firstname, surname, email, phone, bio, password_hash, created_at, updated_at
	FROM users WHERE login=$1`
	row := r.db.QueryRow(query, login)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Login, &user.Firstname, &user.Surname, &user.Email, &user.Phone, &user.Bio, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepo) GetByID(id string) (*models.User, error) {
	query := `
	SELECT id, login, firstname, surname, email, phone, bio, password_hash, created_at, updated_at
	FROM users WHERE id=$1`
	row := r.db.QueryRow(query, id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Login, &user.Firstname, &user.Surname, &user.Email, &user.Phone, &user.Bio, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepo) Update(user *models.User) error {
	query := `
	UPDATE users SET email=$1, firstname=$2, surname=$3, phone=$4, bio=$5, updated_at=now() WHERE id=$6`
	result, err := r.db.Exec(query, user.Email, user.Firstname, user.Surname, user.Phone, user.Bio, user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *postgresUserRepo) Delete(id string) error {
	query := `DELETE FROM users WHERE id=$1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
