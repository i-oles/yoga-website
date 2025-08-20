package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/pkg/optional"
)

type PendingOperationsRepo struct {
	db       *sql.DB
	collName string
}

func NewPendingOperationsRepo(db *sql.DB) *PendingOperationsRepo {
	return &PendingOperationsRepo{
		db:       db,
		collName: "pending_operations",
	}
}

func (r PendingOperationsRepo) Insert(
	ctx context.Context,
	pendingOperation models.PendingOperation,
) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (id, class_id, operation, email, first_name, last_name, auth_token, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);", r.collName)

	_, err := r.db.ExecContext(
		ctx,
		query,
		pendingOperation.ID,
		pendingOperation.ClassID,
		pendingOperation.Operation,
		pendingOperation.Email,
		pendingOperation.FirstName,
		pendingOperation.LastName,
		pendingOperation.AuthToken,
		pendingOperation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pending confirmation: %w", err)
	}

	return nil
}

func (r PendingOperationsRepo) Get(ctx context.Context, token string) (optional.Optional[models.PendingOperation], error) {
	var operation models.PendingOperation

	query := fmt.Sprintf("SELECT id, class_id, operation, email, first_name, last_name, auth_token, created_at FROM %s WHERE auth_token = $1;", r.collName)

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&operation.ID,
		&operation.ClassID,
		&operation.Operation,
		&operation.Email,
		&operation.FirstName,
		&operation.LastName,
		&operation.AuthToken,
		&operation.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return optional.Empty[models.PendingOperation](), fmt.Errorf("pending operation does not exists in database: %w", err)
		}

		return optional.Empty[models.PendingOperation](), fmt.Errorf("failed while getting pending operation: %w", err)
	}

	return optional.Of(operation), nil
}

// TODO: here should delete(ctx, id) not token ??
func (r PendingOperationsRepo) Delete(ctx context.Context, authToken string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE auth_token = $1;", r.collName)

	_, err := r.db.ExecContext(ctx, query, authToken)
	if err != nil {
		return fmt.Errorf("delete pending confirmation: %w", err)
	}

	return nil
}
