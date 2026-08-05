package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpacesRepoPostgres struct {
	pool *pgxpool.Pool
}

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

//go:embed sql/spaces/create_space.sql
var createSpaceSQL string

//go:embed sql/spaces/get_space_by_id.sql
var getSpaceByIDSQL string

//go:embed sql/spaces/get_space_by_slug.sql
var getSpaceBySlugSQL string

//go:embed sql/spaces/list_spaces.sql
var listSpacesSQL string

//go:embed sql/spaces/update_space.sql
var updateSpaceSQL string

//go:embed sql/spaces/archive_space.sql
var archiveSpaceSQL string

//go:embed sql/spaces/add_member.sql
var addMemberSQL string

//go:embed sql/spaces/list_members.sql
var listMembersSQL string

//go:embed sql/spaces/count_members.sql
var countMembersSQL string

//go:embed sql/spaces/is_member.sql
var isMemberSQL string

//go:embed sql/spaces/is_owner_or_admin.sql
var isOwnerOrAdminSQL string

//go:embed sql/spaces/replace_stack.sql
var replaceStackSQL string

//go:embed sql/spaces/add_stack_item.sql
var addStackItemSQL string

//go:embed sql/spaces/list_stack.sql
var listStackSQL string

//go:embed sql/spaces/replace_links.sql
var replaceLinksSQL string

//go:embed sql/spaces/add_link.sql
var addLinkSQL string

//go:embed sql/spaces/list_links.sql
var listLinksSQL string

func NewSpacesRepoPostgres(pool *pgxpool.Pool) *SpacesRepoPostgres {
	return &SpacesRepoPostgres{pool: pool}
}

func (r *SpacesRepoPostgres) Create(ctx context.Context, space domain.Space, stack []domain.SpaceStackItem, links []domain.SpaceLink) (domain.Space, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Space{}, fmt.Errorf("begin create space transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	created, err := createSpace(ctx, tx, space)
	if err != nil {
		return domain.Space{}, err
	}
	if err := addMember(ctx, tx, domain.SpaceMember{
		SpaceID: created.ID,
		UserID:  created.OwnerID,
		Role:    domain.MemberRoleOwner,
	}); err != nil {
		return domain.Space{}, err
	}
	if err := replaceStack(ctx, tx, created.ID, stack); err != nil {
		return domain.Space{}, err
	}
	if err := replaceLinks(ctx, tx, created.ID, links); err != nil {
		return domain.Space{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Space{}, fmt.Errorf("commit create space transaction: %w", err)
	}
	return created, nil
}

func (r *SpacesRepoPostgres) GetByID(ctx context.Context, id domain.SpaceID) (domain.Space, error) {
	return scanSpace(r.pool.QueryRow(ctx, getSpaceByIDSQL, id))
}

func (r *SpacesRepoPostgres) GetBySlug(ctx context.Context, slug string) (domain.Space, error) {
	return scanSpace(r.pool.QueryRow(ctx, getSpaceBySlugSQL, slug))
}

func (r *SpacesRepoPostgres) List(ctx context.Context, filter domain.SpaceListFilter) ([]domain.Space, error) {
	var viewerID any
	if filter.ViewerID != nil {
		viewerID = int64(*filter.ViewerID)
	}
	rows, err := r.pool.Query(ctx, listSpacesSQL, filter.Status, filter.Visibility, viewerID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("list spaces: %w", err)
	}
	defer rows.Close()

	spaces := make([]domain.Space, 0)
	for rows.Next() {
		space, err := scanSpace(rows)
		if err != nil {
			return nil, err
		}
		spaces = append(spaces, space)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate spaces: %w", err)
	}
	return spaces, nil
}

func (r *SpacesRepoPostgres) Update(ctx context.Context, space domain.Space, stack *[]domain.SpaceStackItem, links *[]domain.SpaceLink) (domain.Space, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Space{}, fmt.Errorf("begin update space transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	updated, err := scanSpace(tx.QueryRow(
		ctx,
		updateSpaceSQL,
		space.ID,
		space.Title,
		space.ShortDescription,
		space.Description,
		space.WhyItMatters,
		space.Status,
		space.Visibility,
		space.Stage,
		space.CoverIcon,
		space.AccentColor,
	))
	if err != nil {
		return domain.Space{}, err
	}
	if stack != nil {
		if err := replaceStack(ctx, tx, space.ID, *stack); err != nil {
			return domain.Space{}, err
		}
	}
	if links != nil {
		if err := replaceLinks(ctx, tx, space.ID, *links); err != nil {
			return domain.Space{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Space{}, fmt.Errorf("commit update space transaction: %w", err)
	}
	return updated, nil
}

func (r *SpacesRepoPostgres) Archive(ctx context.Context, slug string, archivedAt time.Time) (domain.Space, error) {
	return scanSpace(r.pool.QueryRow(ctx, archiveSpaceSQL, slug, archivedAt))
}

func (r *SpacesRepoPostgres) AddMember(ctx context.Context, member domain.SpaceMember) error {
	return addMember(ctx, r.pool, member)
}

func (r *SpacesRepoPostgres) ListMembers(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceMember, error) {
	rows, err := r.pool.Query(ctx, listMembersSQL, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list space members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.SpaceMember, 0)
	for rows.Next() {
		var member domain.SpaceMember
		var role string
		if err := rows.Scan(
			&member.SpaceID,
			&member.UserID,
			&member.Username,
			&role,
			&member.Title,
			&member.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan space member: %w", err)
		}
		member.Role = domain.MemberRole(role)
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate space members: %w", err)
	}
	return members, nil
}

func (r *SpacesRepoPostgres) CountMembers(ctx context.Context, spaceID domain.SpaceID) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, countMembersSQL, spaceID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count space members: %w", err)
	}
	return count, nil
}

func (r *SpacesRepoPostgres) IsMember(ctx context.Context, spaceID domain.SpaceID, userID domain.UserID) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, isMemberSQL, spaceID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check space membership: %w", err)
	}
	return exists, nil
}

func (r *SpacesRepoPostgres) IsOwnerOrAdmin(ctx context.Context, spaceID domain.SpaceID, userID domain.UserID) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, isOwnerOrAdminSQL, spaceID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check space management permission: %w", err)
	}
	return exists, nil
}

func (r *SpacesRepoPostgres) ReplaceStack(ctx context.Context, spaceID domain.SpaceID, stack []domain.SpaceStackItem) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin replace stack transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := replaceStack(ctx, tx, spaceID, stack); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace stack transaction: %w", err)
	}
	return nil
}

func (r *SpacesRepoPostgres) ListStack(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceStackItem, error) {
	rows, err := r.pool.Query(ctx, listStackSQL, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list space stack: %w", err)
	}
	defer rows.Close()

	stack := make([]domain.SpaceStackItem, 0)
	for rows.Next() {
		var item domain.SpaceStackItem
		if err := rows.Scan(&item.ID, &item.SpaceID, &item.Name, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan space stack item: %w", err)
		}
		stack = append(stack, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate space stack: %w", err)
	}
	return stack, nil
}

func (r *SpacesRepoPostgres) ReplaceLinks(ctx context.Context, spaceID domain.SpaceID, links []domain.SpaceLink) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin replace links transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := replaceLinks(ctx, tx, spaceID, links); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace links transaction: %w", err)
	}
	return nil
}

func (r *SpacesRepoPostgres) ListLinks(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceLink, error) {
	rows, err := r.pool.Query(ctx, listLinksSQL, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list space links: %w", err)
	}
	defer rows.Close()

	links := make([]domain.SpaceLink, 0)
	for rows.Next() {
		var link domain.SpaceLink
		var linkType string
		if err := rows.Scan(
			&link.ID,
			&link.SpaceID,
			&linkType,
			&link.Title,
			&link.URL,
			&link.Position,
			&link.CreatedAt,
			&link.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan space link: %w", err)
		}
		link.Type = domain.LinkType(linkType)
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate space links: %w", err)
	}
	return links, nil
}

func createSpace(ctx context.Context, db dbtx, space domain.Space) (domain.Space, error) {
	return scanSpace(db.QueryRow(
		ctx,
		createSpaceSQL,
		space.OwnerID,
		space.Slug,
		space.Title,
		space.ShortDescription,
		space.Description,
		space.WhyItMatters,
		space.Status,
		space.Visibility,
		space.Stage,
		space.CoverIcon,
		space.AccentColor,
	))
}

func addMember(ctx context.Context, db dbtx, member domain.SpaceMember) error {
	if _, err := db.Exec(ctx, addMemberSQL, member.SpaceID, member.UserID, member.Role, member.Title); err != nil {
		return mapPostgresError(err)
	}
	return nil
}

func replaceStack(ctx context.Context, db dbtx, spaceID domain.SpaceID, stack []domain.SpaceStackItem) error {
	if _, err := db.Exec(ctx, replaceStackSQL, spaceID); err != nil {
		return fmt.Errorf("clear space stack: %w", err)
	}
	for _, item := range stack {
		if _, err := db.Exec(ctx, addStackItemSQL, spaceID, item.Name); err != nil {
			return mapPostgresError(err)
		}
	}
	return nil
}

func replaceLinks(ctx context.Context, db dbtx, spaceID domain.SpaceID, links []domain.SpaceLink) error {
	if _, err := db.Exec(ctx, replaceLinksSQL, spaceID); err != nil {
		return fmt.Errorf("clear space links: %w", err)
	}
	for _, link := range links {
		if _, err := db.Exec(ctx, addLinkSQL, spaceID, link.Type, link.Title, link.URL, link.Position); err != nil {
			return mapPostgresError(err)
		}
	}
	return nil
}

func scanSpace(row pgx.Row) (domain.Space, error) {
	var space domain.Space
	var status, visibility, stage, coverIcon, accentColor string
	if err := row.Scan(
		&space.ID,
		&space.OwnerID,
		&space.Slug,
		&space.Title,
		&space.ShortDescription,
		&space.Description,
		&space.WhyItMatters,
		&status,
		&visibility,
		&stage,
		&coverIcon,
		&accentColor,
		&space.CreatedAt,
		&space.UpdatedAt,
		&space.ArchivedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Space{}, domain.ErrSpaceNotFound
		}
		return domain.Space{}, mapPostgresError(err)
	}
	space.Status = domain.SpaceStatus(status)
	space.Visibility = domain.SpaceVisibility(visibility)
	space.Stage = domain.SpaceStage(stage)
	space.CoverIcon = domain.CoverIcon(coverIcon)
	space.AccentColor = domain.AccentColor(accentColor)
	return space, nil
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "ux_spaces_slug" {
				return domain.ErrSlugAlreadyExists
			}
			return domain.ErrInvalidInput
		case "23503", "23514":
			return domain.ErrInvalidInput
		}
	}
	return fmt.Errorf("postgres spaces repository: %w", err)
}
