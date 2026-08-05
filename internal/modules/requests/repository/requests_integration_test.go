package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestListBySpaceIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	seed := time.Now().UnixNano()
	spaceOwnerID := insertRequestTestUser(t, ctx, tx, seed, "space_owner")
	requestOwnerID := insertRequestTestUser(t, ctx, tx, seed, "request_owner")
	viewerID := insertRequestTestUser(t, ctx, tx, seed, "viewer")
	respondentA := insertRequestTestUser(t, ctx, tx, seed, "ra")
	respondentB := insertRequestTestUser(t, ctx, tx, seed, "rb")

	spaceID := insertRequestTestSpace(t, ctx, tx, spaceOwnerID, fmt.Sprintf("requests-%d", seed))
	otherSpaceID := insertRequestTestSpace(t, ctx, tx, spaceOwnerID, fmt.Sprintf("other-%d", seed))

	baseTime := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	publicID := insertRequestTestRow(t, ctx, tx, spaceID, requestOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityPublic, "Public request", baseTime)
	membersID := insertRequestTestRow(t, ctx, tx, spaceID, requestOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityMembers, "Members request", baseTime)
	registeredID := insertRequestTestRow(t, ctx, tx, spaceID, requestOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityRegistered, "Registered request", baseTime.Add(time.Minute))
	privateID := insertRequestTestRow(t, ctx, tx, spaceID, spaceOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityPrivate, "Private request", baseTime.Add(2*time.Minute))
	closedID := insertRequestTestRow(t, ctx, tx, spaceID, spaceOwnerID, domain.RequestStatusClosed, domain.RequestVisibilityPublic, "Closed request", baseTime)
	draftID := insertRequestTestRow(t, ctx, tx, spaceID, requestOwnerID, domain.RequestStatusDraft, domain.RequestVisibilityPrivate, "Draft request", baseTime)
	archivedID := insertRequestTestRow(t, ctx, tx, spaceID, spaceOwnerID, domain.RequestStatusArchived, domain.RequestVisibilityPrivate, "Archived request", baseTime)
	_ = insertRequestTestRow(t, ctx, tx, otherSpaceID, spaceOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityPublic, "Other space request", baseTime.Add(3*time.Minute))
	deletedID := insertRequestTestRow(t, ctx, tx, spaceID, spaceOwnerID, domain.RequestStatusOpen, domain.RequestVisibilityPublic, "Deleted request", baseTime.Add(4*time.Minute))
	if _, err := tx.Exec(ctx, `UPDATE requests SET deleted_at = now() WHERE id = $1`, deletedID); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE requests SET expected_commitment = NULL, published_at = NULL, closed_at = NULL WHERE id = $1`, publicID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO request_responses (request_id, respondent_id, message) VALUES ($1, $2, 'First'), ($1, $3, 'Second')`, publicID, respondentA, respondentB); err != nil {
		t.Fatal(err)
	}

	t.Run("anonymous sees public and factual response count without duplicates", func(t *testing.T) {
		page, err := listBySpace(ctx, tx, domain.ListBySpaceParams{SpaceID: spaceID, Status: domain.RequestStatusOpen, Limit: 20})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != publicID {
			t.Fatalf("page = %+v", page)
		}
		if page.Items[0].ResponsesCount != 2 {
			t.Fatalf("responses_count = %d, want 2", page.Items[0].ResponsesCount)
		}
		if page.Items[0].Type != domain.RequestTypeLookingForContributor ||
			page.Items[0].Status != domain.RequestStatusOpen ||
			page.Items[0].Visibility != domain.RequestVisibilityPublic ||
			page.Items[0].Format != domain.RequestFormatPortfolio {
			t.Fatalf("enum mapping = %+v", page.Items[0])
		}
		if page.Items[0].ExpectedCommitment != nil || page.Items[0].PublishedAt != nil || page.Items[0].ClosedAt != nil {
			t.Fatalf("nullable fields = %+v", page.Items[0])
		}
	})

	t.Run("authenticated non-member sees registered but not members", func(t *testing.T) {
		page, err := listBySpace(ctx, tx, domain.ListBySpaceParams{
			SpaceID: spaceID, ViewerID: &viewerID, Status: domain.RequestStatusOpen, Limit: 20,
		})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 2 || len(page.Items) != 2 || page.Items[0].ID != registeredID || page.Items[1].ID != publicID {
			t.Fatalf("page = %+v", page)
		}
	})

	t.Run("member visibility and deterministic pagination", func(t *testing.T) {
		page, err := listBySpace(ctx, tx, domain.ListBySpaceParams{
			SpaceID: spaceID, ViewerID: &viewerID, IsMember: true, Status: domain.RequestStatusOpen, Limit: 2, Offset: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 3 || len(page.Items) != 2 || page.Items[0].ID != membersID || page.Items[1].ID != publicID {
			t.Fatalf("page = %+v", page)
		}
	})

	t.Run("request owner sees own private draft", func(t *testing.T) {
		page, err := listBySpace(ctx, tx, domain.ListBySpaceParams{
			SpaceID: spaceID, ViewerID: &requestOwnerID, Status: domain.RequestStatusDraft, Limit: 20,
		})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != draftID {
			t.Fatalf("page = %+v", page)
		}
	})

	t.Run("space owner or admin flag sees all visibility and lifecycle states", func(t *testing.T) {
		openPage, err := listBySpace(ctx, tx, domain.ListBySpaceParams{
			SpaceID: spaceID, ViewerID: &spaceOwnerID, IsMember: true, CanManage: true, Status: domain.RequestStatusOpen, Limit: 20,
		})
		if err != nil {
			t.Fatal(err)
		}
		if openPage.Total != 4 || len(openPage.Items) != 4 || openPage.Items[0].ID != privateID || openPage.Items[1].ID != registeredID {
			t.Fatalf("open page = %+v", openPage)
		}

		for status, wantID := range map[domain.RequestStatus]int64{
			domain.RequestStatusClosed: closedID, domain.RequestStatusArchived: archivedID,
		} {
			page, err := listBySpace(ctx, tx, domain.ListBySpaceParams{
				SpaceID: spaceID, ViewerID: &spaceOwnerID, IsMember: true, CanManage: true, Status: status, Limit: 20,
			})
			if err != nil {
				t.Fatal(err)
			}
			if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != wantID {
				t.Fatalf("status %s page = %+v", status, page)
			}
		}
	})

	t.Run("unknown Space ID and empty status return an empty page", func(t *testing.T) {
		for _, params := range []domain.ListBySpaceParams{
			{SpaceID: 1<<62 - 1, Status: domain.RequestStatusOpen, Limit: 20},
			{SpaceID: spaceID, Status: domain.RequestStatusArchived, Limit: 20},
		} {
			page, err := listBySpace(ctx, tx, params)
			if err != nil {
				t.Fatal(err)
			}
			if page.Total != 0 || page.Items == nil || len(page.Items) != 0 {
				t.Fatalf("page = %+v", page)
			}
		}
	})
}

func insertRequestTestUser(t *testing.T, ctx context.Context, tx pgx.Tx, seed int64, suffix string) int64 {
	t.Helper()
	var id int64
	username := fmt.Sprintf("%s_%d", suffix, seed)
	email := fmt.Sprintf("%s_%d@example.com", suffix, seed)
	if err := tx.QueryRow(ctx, `INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id`, username, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertRequestTestSpace(t *testing.T, ctx context.Context, tx pgx.Tx, ownerID int64, slug string) int64 {
	t.Helper()
	var id int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO spaces (
			owner_id, slug, title, short_description, status, visibility, stage, cover_icon, accent_color
		) VALUES ($1, $2, 'Requests integration', 'A Space used for request repository tests.', 'active', 'public', 'mvp', 'other', 'gray')
		RETURNING id
	`, ownerID, slug).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertRequestTestRow(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	spaceID int64,
	ownerID int64,
	status domain.RequestStatus,
	visibility domain.RequestVisibility,
	title string,
	createdAt time.Time,
) int64 {
	t.Helper()
	var id int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO requests (
			space_id, owner_id, type, status, visibility, title, description, format, created_at, updated_at
		) VALUES ($1, $2, 'looking_for_contributor', $3, $4, $5, 'Request integration description.', 'portfolio', $6, $6)
		RETURNING id
	`, spaceID, ownerID, status, visibility, title, createdAt).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
