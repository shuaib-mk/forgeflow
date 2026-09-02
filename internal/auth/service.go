package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/forgeflow/forgeflow/pkg/validation"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const SessionLifetime = 24 * time.Hour

type Repository interface {
	Register(context.Context, models.User, models.Organization) error
	UserByEmail(context.Context, string) (models.User, error)
	UserByID(context.Context, models.ID) (models.User, error)
	CreateSession(context.Context, string, string, models.ID, time.Time) error
	UserBySession(context.Context, string) (models.User, error)
	DeleteSession(context.Context, string) error
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

type RegisterInput struct {
	Email            string `json:"email"`
	DisplayName      string `json:"displayName"`
	Password         string `json:"password"`
	OrganizationName string `json:"organizationName"`
	OrganizationSlug string `json:"organizationSlug"`
}
type Registration struct {
	User         models.User         `json:"user"`
	Organization models.Organization `json:"organization"`
}
type Session struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expiresAt"`
	User      models.User `json:"user"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (Registration, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if !validation.Email(email) {
		return Registration{}, domain.Invalid("email", "must be a valid email address")
	}
	if !validation.Required(input.DisplayName, 100) {
		return Registration{}, domain.Invalid("displayName", "must contain 1 to 100 characters")
	}
	if len(input.Password) < 12 || len(input.Password) > 128 {
		return Registration{}, domain.Invalid("password", "must contain 12 to 128 characters")
	}
	if !validation.Required(input.OrganizationName, 100) {
		return Registration{}, domain.Invalid("organizationName", "must contain 1 to 100 characters")
	}
	if !validation.Slug(input.OrganizationSlug) {
		return Registration{}, domain.Invalid("organizationSlug", "must be a lowercase URL-safe slug")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Registration{}, err
	}
	now := s.now().UTC()
	user := models.User{ID: models.ID(uuid.NewString()), Email: email, DisplayName: strings.TrimSpace(input.DisplayName), PasswordHash: string(hash), CreatedAt: now, UpdatedAt: now}
	organization := models.Organization{ID: models.ID(uuid.NewString()), Name: strings.TrimSpace(input.OrganizationName), Slug: input.OrganizationSlug, CreatedAt: now}
	if err := s.repository.Register(ctx, user, organization); err != nil {
		return Registration{}, err
	}
	user.PasswordHash = ""
	return Registration{User: user, Organization: organization}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.repository.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return Session{}, domain.ErrUnauthorized
		}
		return Session{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return Session{}, domain.ErrUnauthorized
	}
	return s.newSession(ctx, user)
}

func (s *Service) Authenticate(ctx context.Context, token string) (models.User, error) {
	if token == "" {
		return models.User{}, domain.ErrUnauthorized
	}
	return s.repository.UserBySession(ctx, hashToken(token))
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.repository.DeleteSession(ctx, hashToken(token))
}

func (s *Service) newSession(ctx context.Context, user models.User) (Session, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	expires := s.now().UTC().Add(SessionLifetime)
	if err := s.repository.CreateSession(ctx, uuid.NewString(), hashToken(token), user.ID, expires); err != nil {
		return Session{}, err
	}
	user.PasswordHash = ""
	return Session{Token: token, ExpiresAt: expires, User: user}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
