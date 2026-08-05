package requestshttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/requests/delivery/transport"
	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	"github.com/butorovv/worknet-backend/internal/modules/requests/usecase"
)

func TestListHandlerSuccess(t *testing.T) {
	createdAt := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	uc := &handlerRequestsUseCase{
		result: usecase.ListSpaceRequestsResult{
			Items: []domain.Request{{
				ID:             11,
				SpaceID:        2,
				OwnerID:        3,
				Type:           domain.RequestTypeLookingForContributor,
				Status:         domain.RequestStatusClosed,
				Visibility:     domain.RequestVisibilityPublic,
				Title:          "Backend contributor",
				Description:    "Help implement the public API.",
				Format:         domain.RequestFormatPortfolio,
				ResponsesCount: 4,
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
			}},
			Pagination: usecase.Pagination{Limit: 5, Offset: 2, Total: 9},
		},
	}
	handler := NewRequestsHandler(uc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spaces/worknet/requests?status=closed&limit=5&offset=2", nil)
	req.SetPathValue("slug", "worknet")
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if uc.lastQuery.Slug != "worknet" || uc.lastQuery.Status != domain.RequestStatusClosed || uc.lastQuery.Limit != 5 || uc.lastQuery.Offset != 2 || uc.lastQuery.ViewerID != nil {
		t.Fatalf("query = %+v", uc.lastQuery)
	}
	var response transport.SpaceRequestsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 || response.Items[0].ResponsesCount != 4 || response.Pagination.Total != 9 {
		t.Fatalf("response = %+v", response)
	}
}

func TestListHandlerWritesEmptyItemsArrayAndDefaults(t *testing.T) {
	uc := &handlerRequestsUseCase{result: usecase.ListSpaceRequestsResult{
		Items:      []domain.Request{},
		Pagination: usecase.Pagination{Limit: 20},
	}}
	handler := NewRequestsHandler(uc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spaces/worknet/requests", nil)
	req.SetPathValue("slug", "worknet")
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if uc.lastQuery.Status != domain.RequestStatusOpen || uc.lastQuery.Limit != 20 || uc.lastQuery.Offset != 0 {
		t.Fatalf("default query = %+v", uc.lastQuery)
	}
}

func TestListHandlerRejectsInvalidQuery(t *testing.T) {
	for _, rawQuery := range []string{
		"limit=nope",
		"limit=0",
		"limit=-1",
		"limit=101",
		"offset=nope",
		"offset=-1",
		"status=pending",
	} {
		t.Run(rawQuery, func(t *testing.T) {
			uc := &handlerRequestsUseCase{}
			handler := NewRequestsHandler(uc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/spaces/worknet/requests?"+rawQuery, nil)
			req.SetPathValue("slug", "worknet")
			rec := httptest.NewRecorder()

			handler.List(rec, req)

			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"space_invalid_input"`) {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if uc.calls != 0 {
				t.Fatal("use case must not be called")
			}
		})
	}
}

func TestListHandlerMapsErrorsWithoutInternalLeak(t *testing.T) {
	databaseErr := errors.New("pq: secret database detail")
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid input", err: domain.ErrInvalidInput, wantStatus: http.StatusBadRequest, wantCode: "space_invalid_input"},
		{name: "not found", err: domain.ErrSpaceNotFound, wantStatus: http.StatusNotFound, wantCode: "space_not_found"},
		{name: "forbidden", err: domain.ErrSpaceForbidden, wantStatus: http.StatusForbidden, wantCode: "space_forbidden"},
		{name: "archived", err: domain.ErrSpaceArchived, wantStatus: http.StatusGone, wantCode: "space_archived"},
		{name: "internal", err: databaseErr, wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &handlerRequestsUseCase{err: tt.err}
			handler := NewRequestsHandler(uc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/spaces/worknet/requests", nil)
			req.SetPathValue("slug", "worknet")
			rec := httptest.NewRecorder()

			handler.List(rec, req)

			if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), "secret database detail") {
				t.Fatalf("internal detail leaked: %s", rec.Body.String())
			}
		})
	}
}

func TestListHandlerRejectsWrongMethod(t *testing.T) {
	uc := &handlerRequestsUseCase{}
	handler := NewRequestsHandler(uc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spaces/worknet/requests", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusMethodNotAllowed || uc.calls != 0 {
		t.Fatalf("status = %d, calls = %d", rec.Code, uc.calls)
	}
}

type handlerRequestsUseCase struct {
	result    usecase.ListSpaceRequestsResult
	err       error
	lastQuery usecase.ListSpaceRequestsQuery
	calls     int
}

func (f *handlerRequestsUseCase) ListSpaceRequests(_ context.Context, query usecase.ListSpaceRequestsQuery) (usecase.ListSpaceRequestsResult, error) {
	f.calls++
	f.lastQuery = query
	return f.result, f.err
}
