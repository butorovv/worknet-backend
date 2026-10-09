package app

import (
	"context"
	"errors"
	"testing"

	requestsdomain "github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	spacesdomain "github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	spacesusecase "github.com/butorovv/worknet-backend/internal/modules/spaces/usecase"
)

func TestRequestsSpaceAccessAdapterMapsViewerAndFlags(t *testing.T) {
	viewerID := int64(42)
	spaces := &fakeSpacesAccessService{access: spacesusecase.SpaceAccess{
		SpaceID: 7, IsMember: true, IsOwner: false, IsAdmin: true,
	}}
	adapter := newRequestsSpaceAccessAdapter(spaces)

	access, err := adapter.ResolveAccessibleSpace(context.Background(), "worknet", &viewerID)
	if err != nil {
		t.Fatal(err)
	}
	if access.SpaceID != 7 || !access.IsMember || access.IsOwner || !access.IsAdmin {
		t.Fatalf("access = %+v", access)
	}
	if spaces.query.Slug != "worknet" || !spaces.query.Viewer.IsAuthenticated || spaces.query.Viewer.UserID != 42 {
		t.Fatalf("query = %+v", spaces.query)
	}
}

func TestRequestsSpaceAccessAdapterMapsAnonymousViewer(t *testing.T) {
	spaces := &fakeSpacesAccessService{access: spacesusecase.SpaceAccess{SpaceID: 7}}
	adapter := newRequestsSpaceAccessAdapter(spaces)

	_, err := adapter.ResolveAccessibleSpace(context.Background(), "worknet", nil)
	if err != nil {
		t.Fatal(err)
	}
	if spaces.query.Viewer.IsAuthenticated || spaces.query.Viewer.UserID != 0 {
		t.Fatalf("viewer = %+v", spaces.query.Viewer)
	}
}

func TestRequestsSpaceAccessAdapterMapsSpaceErrors(t *testing.T) {
	infrastructureErr := errors.New("postgres unavailable")
	tests := []struct {
		spaceErr error
		wantErr  error
	}{
		{spaceErr: spacesdomain.ErrInvalidInput, wantErr: requestsdomain.ErrInvalidInput},
		{spaceErr: spacesdomain.ErrSpaceNotFound, wantErr: requestsdomain.ErrSpaceNotFound},
		{spaceErr: spacesdomain.ErrForbidden, wantErr: requestsdomain.ErrSpaceForbidden},
		{spaceErr: spacesdomain.ErrSpaceArchived, wantErr: requestsdomain.ErrSpaceArchived},
		{spaceErr: infrastructureErr, wantErr: infrastructureErr},
	}
	for _, tt := range tests {
		adapter := newRequestsSpaceAccessAdapter(&fakeSpacesAccessService{err: tt.spaceErr})
		_, err := adapter.ResolveAccessibleSpace(context.Background(), "worknet", nil)
		if !errors.Is(err, tt.wantErr) {
			t.Fatalf("space error %v mapped to %v, want %v", tt.spaceErr, err, tt.wantErr)
		}
	}
}

type fakeSpacesAccessService struct {
	access spacesusecase.SpaceAccess
	err    error
	query  spacesusecase.ResolveSpaceAccessQuery
}

func (f *fakeSpacesAccessService) ResolveSpaceAccess(_ context.Context, query spacesusecase.ResolveSpaceAccessQuery) (spacesusecase.SpaceAccess, error) {
	f.query = query
	return f.access, f.err
}
