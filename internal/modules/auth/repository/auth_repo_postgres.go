package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepoPostgres struct {
	pool *pgxpool.Pool
}

//go:embed sql/auth/create_user.sql
var createUserSQL string

//go:embed sql/auth/create_password_identity.sql
var createPasswordIdentitySQL string

//go:embed sql/auth/find_user_by_email.sql
var findUserByEmailSQL string

//go:embed sql/auth/find_user_by_id.sql
var findUserByIDSQL string

//go:embed sql/auth/password_hash_by_user_id.sql
var passwordHashByUserIDSQL string

//go:embed sql/auth/touch_password_identity.sql
var touchPasswordIdentitySQL string

func NewAuthRepoPostgres(pool *pgxpool.Pool) *AuthRepoPostgres {
	return &AuthRepoPostgres{pool: pool}
}

func (r *AuthRepoPostgres) CreatePasswordUser(ctx context.Context, username, email, passwordHash string) (domain.User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var user domain.User
	err = tx.QueryRow(ctx, createUserSQL, username, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, mapPostgresError(err)
	}

	_, err = tx.Exec(ctx, createPasswordIdentitySQL, user.ID, domain.ProviderPassword, passwordHash)
	if err != nil {
		return domain.User{}, mapPostgresError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit create user transaction: %w", err)
	}

	return user, nil
}

func (r *AuthRepoPostgres) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, findUserByEmailSQL, email))
}

func (r *AuthRepoPostgres) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, findUserByIDSQL, id))
}

func (r *AuthRepoPostgres) PasswordHashByUserID(ctx context.Context, userID domain.UserID) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, passwordHashByUserIDSQL, userID, domain.ProviderPassword).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", fmt.Errorf("select password identity: %w", err)
	}
	return hash, nil
}

func (r *AuthRepoPostgres) TouchPasswordIdentity(ctx context.Context, userID domain.UserID, usedAt time.Time) error {
	commandTag, err := r.pool.Exec(ctx, touchPasswordIdentitySQL, userID, usedAt, domain.ProviderPassword)
	if err != nil {
		return fmt.Errorf("touch password identity: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return domain.ErrUserAlreadyExists
		case "23514":
			return domain.ErrInvalidInput
		}
	}
	return fmt.Errorf("postgres auth repository: %w", err)
}
