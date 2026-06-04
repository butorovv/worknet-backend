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

type authUseCase struct {
	repo       AuthRepository
	sessions   SessionStore
	hasher     PasswordHasher
	tokens     SessionTokenGenerator
	clock      Clock
	sessionTTL time.Duration
}

func NewAuthUseCase(
	repo AuthRepository,
	sessions SessionStore,
	hasher PasswordHasher,
	tokens SessionTokenGenerator,
	sessionTTL time.Duration,
) (*authUseCase, error) {
	return newAuthUseCase(repo, sessions, hasher, tokens, RealClock{}, sessionTTL)
}

func newAuthUseCase(
	repo AuthRepository,
	sessions SessionStore,
	hasher PasswordHasher,
	tokens SessionTokenGenerator,
	clock Clock,
	sessionTTL time.Duration,
) (*authUseCase, error) {
	if repo == nil {
		return nil, errors.New("auth repository is required")
	}
	if sessions == nil {
		return nil, errors.New("auth session store is required")
	}
	if hasher == nil {
		return nil, errors.New("auth password hasher is required")
	}
	if tokens == nil {
		return nil, errors.New("auth session token generator is required")
	}
	if clock == nil {
		clock = RealClock{}
	}
	if sessionTTL <= 0 {
		return nil, errors.New("auth session ttl must be positive")
	}

	return &authUseCase{
		repo:       repo,
		sessions:   sessions,
		hasher:     hasher,
		tokens:     tokens,
		clock:      clock,
		sessionTTL: sessionTTL,
	}, nil
}

func (s *authUseCase) Register(ctx context.Context, cmd RegisterCommand) (AuthResult, error) {
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

	user, err := s.repo.CreatePasswordUser(ctx, username, email, passwordHash)
	if err != nil {
		return AuthResult{}, err
	}

	return s.createSession(ctx, user)
}

func (s *authUseCase) Login(ctx context.Context, cmd LoginCommand) (AuthResult, error) {
	email := normalizeEmail(cmd.Email)
	password := cmd.Password

	if !validEmail(email) || password == "" {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return AuthResult{}, domain.ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if !user.IsActive() {
		return AuthResult{}, domain.ErrUserInactive
	}

	passwordHash, err := s.repo.PasswordHashByUserID(ctx, user.ID)
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
	if err := s.repo.TouchPasswordIdentity(ctx, user.ID, now); err != nil {
		return AuthResult{}, err
	}

	return s.createSessionAt(ctx, user, now)
}

func (s *authUseCase) Logout(ctx context.Context, cmd LogoutCommand) error {
	sessionID := strings.TrimSpace(cmd.SessionID)
	if sessionID == "" {
		return domain.ErrSessionNotFound
	}
	return s.sessions.Delete(ctx, sessionID)
}

func (s *authUseCase) ValidateSession(ctx context.Context, cmd ValidateSessionCommand) (CurrentUser, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	if sessionID == "" {
		return CurrentUser{}, domain.ErrSessionNotFound
	}

	session, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return CurrentUser{}, err
	}

	user, err := s.repo.FindByID(ctx, session.UserID)
	if err != nil {
		return CurrentUser{}, err
	}
	if !user.IsActive() {
		return CurrentUser{}, domain.ErrUserInactive
	}

	return currentUserFromDomain(user), nil
}

func (s *authUseCase) createSession(ctx context.Context, user domain.User) (AuthResult, error) {
	return s.createSessionAt(ctx, user, s.clock.Now())
}

func (s *authUseCase) createSessionAt(ctx context.Context, user domain.User, now time.Time) (AuthResult, error) {
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
