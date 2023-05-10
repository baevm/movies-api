package acttokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"movies-api/internal/validator"
	"time"
)

const (
	ScopeActivation    = "activation"
	ScopeAuth          = "authentication"
	ScopePasswordReset = "password-reset"
)

type ActToken struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type ActTokenService struct {
	DB *sql.DB
}

func NewActTokenService(db *sql.DB) *ActTokenService {
	return &ActTokenService{DB: db}
}

func (t ActTokenService) New(userID int64, ttl time.Duration, scope string) (*ActToken, error) {
	token, err := generateActToken(userID, ttl, scope)

	if err != nil {
		return nil, err
	}

	err = t.Create(token)
	return token, err
}

func (t ActTokenService) Create(token *ActToken) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)
	`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)

	return err
}

func (t ActTokenService) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens 
	WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, scope, userID)

	return err
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "Token must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "Token must be 26 characters")
}

func generateActToken(userID int64, ttl time.Duration, scope string) (*ActToken, error) {
	token := &ActToken{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	// fill byte slice with random bytes from OS
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
