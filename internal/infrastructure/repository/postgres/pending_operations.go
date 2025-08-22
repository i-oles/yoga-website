package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/domain/models"

	"github.com/google/uuid"
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
		"INSERT INTO %s (id, class_id, operation, email, first_name, last_name, confirmation_token, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);", r.collName)

	_, err := r.db.ExecContext(
		ctx,
		query,
		pendingOperation.ID,
		pendingOperation.ClassID,
		pendingOperation.Operation,
		pendingOperation.Email,
		pendingOperation.FirstName,
		pendingOperation.LastName,
		pendingOperation.ConfirmationToken,
		pendingOperation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pending confirmation: %w", err)
	}

	return nil
}

func (r PendingOperationsRepo) Get(ctx context.Context, token string) (models.PendingOperation, error) {
	var operation models.PendingOperation

	query := fmt.Sprintf("SELECT id, class_id, operation, email, first_name, last_name, confirmation_token, created_at FROM %s WHERE confirmation_token = $1;", r.collName)

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&operation.ID,
		&operation.ClassID,
		&operation.Operation,
		&operation.Email,
		&operation.FirstName,
		&operation.LastName,
		&operation.ConfirmationToken,
		&operation.CreatedAt,
	)
	if err != nil {
		return models.PendingOperation{}, fmt.Errorf("failed while getting pending operation: %w", err)
	}

	return operation, nil
}

func (r PendingOperationsRepo) CountPendingOperationsPerUser(
	ctx context.Context,
	email string,
	operation models.Operation,
	classID uuid.UUID,
) (int8, error) {
	query := fmt.Sprintf("SELECT COUNT(email) FROM %s WHERE email = $1 AND class_id = $2 AND operation = $3;", r.collName)

	var count int8
	err := r.db.QueryRowContext(ctx, query, email, classID, operation).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed while counting user in class: %w", err)
	}

	return count, nil
}

// TODO: here should delete(ctx, id) not token ??
func (r PendingOperationsRepo) Delete(ctx context.Context, confirmationToken string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE confirmation_token = $1;", r.collName)

	_, err := r.db.ExecContext(ctx, query, confirmationToken)
	if err != nil {
		return fmt.Errorf("delete pending confirmation: %w", err)
	}

	return nil
}
