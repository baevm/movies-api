package permissions

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

type PermissionsService struct {
	db *sql.DB
}

func NewPermissionsService(db *sql.DB) *PermissionsService {
	return &PermissionsService{
		db: db,
	}
}

func (p Permissions) IsInclude(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

func (p PermissionsService) GetAllForUser(userID int64) (Permissions, error) {
	var permissions Permissions

	query := `
	SELECT permissions.code
	FROM permissions
	INNER JOIN users_permissions on users_permissions.permission_id = permissions.id
	INNER JOIN users on users_permissions.user_id = users.id
	WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var perm string

		err := rows.Scan(&perm)

		if err != nil {
			return nil, err
		}

		permissions = append(permissions, perm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (p PermissionsService) AddForUser(userID int64, codes ...string) error {
	query := `
	INSERT INTO users_permissions
	SELECT $1, permissions.id 
	FROM permissions
	WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query, userID, pq.Array(codes))

	return err
}
