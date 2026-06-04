package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]{2,29}$`)

type Dependencies struct {
	Users      UserRepository
	Identities IdentityRepository
	Sessions   SessionStore
	Hasher     PasswordHasher
	Tokens     SessionTokenGenerator
	Clock      Clock
	SessionTTL time.Duration
}

type Service struct {
	users      UserRepository
	identities IdentityRepository
	sessions   SessionStore
	hasher     PasswordHasher
	tokens     SessionTokenGenerator
	clock      Clock
	sessionTTL time.Duration
}

func NewService(deps Dependencies) (*Service, error) {
	if deps.Users == nil {
		return nil, errors.New("auth users repository is required")
	}
	if deps.Identities == nil {
		return nil, errors.New("auth identities repository is required")
	}
	if deps.Sessions == nil {
		return nil, errors.New("auth session store is required")
	}
	if deps.Hasher == nil {
		return nil, errors.New("auth password hasher is required")
	}
	if deps.Tokens == nil {
		return nil, errors.New("auth session token generator is required")
	}
	if deps.Clock == nil {
		deps.Clock = RealClock{}
	}
	if deps.SessionTTL <= 0 {
		return nil, errors.New("auth session ttl must be positive")
	}

	return &Service{
		users:      deps.Users,
		identities: deps.Identities,
		sessions:   deps.Sessions,
		hasher:     deps.Hasher,
		tokens:     deps.Tokens,
		clock:      deps.Clock,
		sessionTTL: deps.SessionTTL,
	}, nil
}

func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (AuthResult, error) {
	username := strings.TrimSpace(cmd.Username)
	email := normalizeEmail(cmd.Email)
	password := cmd.Password

	if !validUsername(username) || !validEmail(email) || !validPassword(password) {
		return AuthResult{}, domain.ErrInvalidInput
	}

	passwordHash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.CreatePasswordUser(ctx, CreatePasswordUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return AuthResult{}, err
	}

	return s.createSession(ctx, user)
}

func (s *Service) Login(ctx context.Context, cmd LoginCommand) (AuthResult, error) {
	email := normalizeEmail(cmd.Email)
	password := cmd.Password

	if !validEmail(email) || password == "" {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return AuthResult{}, domain.ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if !user.IsActive() {
		return AuthResult{}, domain.ErrUserInactive
	}

	passwordHash, err := s.identities.PasswordHashByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return AuthResult{}, domain.ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if err := s.hasher.Compare(ctx, passwordHash, password); err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	now := s.clock.Now()
	if err := s.identities.TouchPasswordIdentity(ctx, user.ID, now); err != nil {
		return AuthResult{}, err
	}

	return s.createSessionAt(ctx, user, now)
}

func (s *Service) Logout(ctx context.Context, cmd LogoutCommand) error {
	sessionID := strings.TrimSpace(cmd.SessionID)
	if sessionID == "" {
		return domain.ErrSessionNotFound
	}
	return s.sessions.Delete(ctx, sessionID)
}

func (s *Service) ValidateSession(ctx context.Context, cmd ValidateSessionCommand) (CurrentUser, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	if sessionID == "" {
		return CurrentUser{}, domain.ErrSessionNotFound
	}

	session, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return CurrentUser{}, err
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return CurrentUser{}, err
	}
	if !user.IsActive() {
		return CurrentUser{}, domain.ErrUserInactive
	}

	return currentUserFromDomain(user), nil
}

func (s *Service) createSession(ctx context.Context, user domain.User) (AuthResult, error) {
	return s.createSessionAt(ctx, user, s.clock.Now())
}

func (s *Service) createSessionAt(ctx context.Context, user domain.User, now time.Time) (AuthResult, error) {
	sessionID, err := s.tokens.Generate(ctx)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate session token: %w", err)
	}

	session := domain.Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User:      currentUserFromDomain(user),
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validUsername(username string) bool {
	return usernamePattern.MatchString(username)
}

func validEmail(email string) bool {
	return strings.Contains(email, "@") && !strings.ContainsAny(email, " \t\r\n")
}

func validPassword(password string) bool {
	return len(password) >= 8
}
