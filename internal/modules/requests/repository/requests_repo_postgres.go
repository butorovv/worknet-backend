package repository

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestsRepoPostgres struct {
	pool *pgxpool.Pool
}

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

//go:embed sql/requests/count_space_requests.sql
var countSpaceRequestsSQL string

//go:embed sql/requests/list_space_requests.sql
var listSpaceRequestsSQL string

func NewRequestsRepoPostgres(pool *pgxpool.Pool) *RequestsRepoPostgres {
	return &RequestsRepoPostgres{pool: pool}
}

func (r *RequestsRepoPostgres) ListBySpace(ctx context.Context, params domain.ListBySpaceParams) (domain.Page, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return domain.Page{}, fmt.Errorf("begin list space requests transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	page, err := listBySpace(ctx, tx, params)
	if err != nil {
		return domain.Page{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Page{}, fmt.Errorf("commit list space requests transaction: %w", err)
	}
	return page, nil
}

func listBySpace(ctx context.Context, db dbtx, params domain.ListBySpaceParams) (domain.Page, error) {
	var viewerID any
	if params.ViewerID != nil {
		viewerID = *params.ViewerID
	}

	var total int
	if err := db.QueryRow(
		ctx,
		countSpaceRequestsSQL,
		params.SpaceID,
		params.Status,
		viewerID,
		params.IsMember,
		params.CanManage,
	).Scan(&total); err != nil {
		return domain.Page{}, fmt.Errorf("count space requests: %w", err)
	}

	rows, err := db.Query(
		ctx,
		listSpaceRequestsSQL,
		params.SpaceID,
		params.Status,
		viewerID,
		params.IsMember,
		params.CanManage,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return domain.Page{}, fmt.Errorf("list space requests: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Request, 0)
	for rows.Next() {
		var request domain.Request
		var requestType, status, visibility, format string
		if err := rows.Scan(
			&request.ID,
			&request.SpaceID,
			&request.OwnerID,
			&requestType,
			&status,
			&visibility,
			&request.Title,
			&request.Description,
			&request.ExpectedCommitment,
			&format,
			&request.ResponsesCount,
			&request.PublishedAt,
			&request.ClosedAt,
			&request.CreatedAt,
			&request.UpdatedAt,
		); err != nil {
			return domain.Page{}, fmt.Errorf("scan space request: %w", err)
		}
		request.Type = domain.RequestType(requestType)
		request.Status = domain.RequestStatus(status)
		request.Visibility = domain.RequestVisibility(visibility)
		request.Format = domain.RequestFormat(format)
		items = append(items, request)
	}
	if err := rows.Err(); err != nil {
		return domain.Page{}, fmt.Errorf("iterate space requests: %w", err)
	}
	return domain.Page{Items: items, Total: total}, nil
}
