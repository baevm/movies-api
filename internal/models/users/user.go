package users

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"movies-api/internal/models"
	"time"
)

type User struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   Password  `json:"-"`
	Created_at time.Time `json:"created_at"`
	Activated  bool      `json:"activated"`
	Version    int       `json:"-"`
}

type UserService struct {
	db *sql.DB
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (u UserService) Create(user *User) error {
	query := `
	INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.Id, &user.Created_at, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (u UserService) GetByEmail(email string) (*User, error) {
	var user User

	query := `
	SELECT id, name, email, password_hash, activated, created_at, version
	FROM users
	WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.db.
		QueryRowContext(ctx, query, user.Email).
		Scan(
			&user.Id,
			&user.Name,
			&user.Email,
			&user.Password.hash,
			&user.Activated,
			&user.Created_at,
			&user.Version,
		)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, models.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserService) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1 
	WHERE id = $5 AND version = $6
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated, user.Id, user.Version}

	err := u.db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return models.ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (u UserService) GetForToken(scope, tokenPlainttext string) (*User, error) {
	hashToken := sha256.Sum256([]byte(tokenPlainttext))

	var user User

	query := `
	SELECT users.id, users.name, users.email, users.password_hash, users.activated, users.created_at, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1 
	AND tokens.scope = $2 
	AND tokens.expiry > $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{hashToken[:], scope, time.Now()}

	err := u.db.QueryRowContext(ctx, query, args...).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Created_at,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, models.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
