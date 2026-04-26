package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password Password `json:"-"`
	IsActive bool     `json:"is_active"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Password struct {
	text *string
	hash []byte
}

func (p *Password) Set(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &password
	p.hash = hash
	return nil
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users(name, email, password) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := u.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Password.hash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, name, email, password, updated_at, created_at FROM users WHERE id = $1`
	row := u.db.QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password.hash, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, name, email, password, updated_at, created_at FROM users WHERE email = $1`
	row := u.db.QueryRowContext(ctx, query, email)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		err := u.CreateUserInvitation(ctx, tx, token, exp, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) CreateUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetUserFromInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Time) (*User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON ui.user_id = u.id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	var user User
	err := tx.QueryRowContext(ctx, query, hashToken, exp).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}
	return &user, nil
}

func (u *UserStore) Update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET name = $1, is_active = $2, updated_at = $3 WHERE id = $4`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Name, user.IsActive, user.UpdatedAt, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (u *UserStore) DeleteUserInvitation(ctx context.Context, user *User) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	res, err := u.db.ExecContext(ctx, query, user.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (u *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		user, err := u.GetUserFromInvitation(ctx, tx, token, time.Now().Add(QueryTimeDuration))
		if err != nil {
			return err
		}
		user.IsActive = true
		if err := u.Update(ctx, tx, user); err != nil {
			return err
		}

		if err := u.DeleteUserInvitation(ctx, user); err != nil {
			return err
		}
		return nil
	})
}

func (u *UserStore) Delete(ctx context.Context, user *User) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		query := `DELETE FROM users WHERE id = $1`
		ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
		defer cancel()

		_, err := tx.ExecContext(ctx, query, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}
