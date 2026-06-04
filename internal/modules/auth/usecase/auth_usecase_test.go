package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
)

func TestRegisterCreatesUserAndSession(t *testing.T) {
	users := newFakeUsers()
	sessions := newFakeSessions()
	uc := newTestUsecase(t, users, sessions)

	result, err := uc.Register(context.Background(), RegisterCommand{
		Username: "georgiy",
		Email:    "GEORGIY@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if result.User.Email != "georgiy@example.com" {
		t.Fatalf("Register() email = %q", result.User.Email)
	}
	if result.SessionID == "" {
		t.Fatal("Register() returned empty session id")
	}
	if _, ok := sessions.byID[result.SessionID]; !ok {
		t.Fatal("Register() did not store session")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	users := newFakeUsers()
	sessions := newFakeSessions()
	uc := newTestUsecase(t, users, sessions)

	_, err := uc.Register(context.Background(), RegisterCommand{
		Username: "georgiy",
		Email:    "georgiy@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err = uc.Login(context.Background(), LoginCommand{
		Email:    "georgiy@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestValidateSessionReturnsCurrentUser(t *testing.T) {
	users := newFakeUsers()
	sessions := newFakeSessions()
	uc := newTestUsecase(t, users, sessions)

	result, err := uc.Register(context.Background(), RegisterCommand{
		Username: "georgiy",
		Email:    "georgiy@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, err := uc.ValidateSession(context.Background(), ValidateSessionCommand{
		SessionID: result.SessionID,
	})
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}
	if user.ID != result.User.ID {
		t.Fatalf("ValidateSession() user id = %d, want %d", user.ID, result.User.ID)
	}
}

func newTestUsecase(t *testing.T, users *fakeUsers, sessions *fakeSessions) *AuthUsecase {
	t.Helper()

	uc, err := newAuthUsecase(
		users,
		sessions,
		fakeHasher{},
		&fakeTokens{},
		fakeClock{now: time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)},
		30*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("NewAuthUsecase() error = %v", err)
	}
	return uc
}

type fakeUsers struct {
	nextID   int64
	byID     map[domain.UserID]domain.User
	byEmail  map[string]domain.UserID
	hashByID map[domain.UserID]string
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		nextID:   1,
		byID:     make(map[domain.UserID]domain.User),
		byEmail:  make(map[string]domain.UserID),
		hashByID: make(map[domain.UserID]string),
	}
}

func (f *fakeUsers) CreatePasswordUser(_ context.Context, username, email, passwordHash string) (domain.User, error) {
	if _, ok := f.byEmail[email]; ok {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	id := domain.UserID(f.nextID)
	f.nextID++
	user := domain.User{
		ID:        id,
		Username:  username,
		Email:     email,
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.byID[id] = user
	f.byEmail[email] = id
	f.hashByID[id] = passwordHash
	return user, nil
}

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (domain.User, error) {
	id, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return f.FindByID(context.Background(), id)
}

func (f *fakeUsers) FindByID(_ context.Context, id domain.UserID) (domain.User, error) {
	user, ok := f.byID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (f *fakeUsers) PasswordHashByUserID(_ context.Context, userID domain.UserID) (string, error) {
	hash, ok := f.hashByID[userID]
	if !ok {
		return "", domain.ErrUserNotFound
	}
	return hash, nil
}

func (f *fakeUsers) TouchPasswordIdentity(_ context.Context, userID domain.UserID, _ time.Time) error {
	if _, ok := f.hashByID[userID]; !ok {
		return domain.ErrUserNotFound
	}
	return nil
}

type fakeSessions struct {
	byID map[string]domain.Session
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{byID: make(map[string]domain.Session)}
}

func (f *fakeSessions) Create(_ context.Context, session domain.Session) error {
	f.byID[session.ID] = session
	return nil
}

func (f *fakeSessions) Get(_ context.Context, sessionID string) (domain.Session, error) {
	session, ok := f.byID[sessionID]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (f *fakeSessions) Delete(_ context.Context, sessionID string) error {
	delete(f.byID, sessionID)
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(_ context.Context, password string) (string, error) {
	return "hash:" + password, nil
}

func (fakeHasher) Compare(_ context.Context, hash, password string) error {
	if hash != "hash:"+password {
		return errors.New("password mismatch")
	}
	return nil
}

type fakeTokens struct {
	next int
}

func (f *fakeTokens) Generate(_ context.Context) (string, error) {
	f.next++
	return "session-token", nil
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
