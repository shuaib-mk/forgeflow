package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type memoryRepository struct {
	user      models.User
	tokenHash string
}

func (r *memoryRepository) Register(_ context.Context, u models.User, _ models.Organization) error {
	r.user = u
	return nil
}
func (r *memoryRepository) UserByEmail(_ context.Context, email string) (models.User, error) {
	if r.user.Email != email {
		return models.User{}, domain.ErrNotFound
	}
	return r.user, nil
}
func (r *memoryRepository) UserByID(context.Context, models.ID) (models.User, error) {
	return r.user, nil
}
func (r *memoryRepository) CreateSession(_ context.Context, _ string, hash string, _ models.ID, _ time.Time) error {
	r.tokenHash = hash
	return nil
}
func (r *memoryRepository) UserBySession(_ context.Context, hash string) (models.User, error) {
	if hash != r.tokenHash {
		return models.User{}, domain.ErrUnauthorized
	}
	return r.user, nil
}
func (r *memoryRepository) DeleteSession(context.Context, string) error { return nil }

func TestRegisterLoginAuthenticate(t *testing.T) {
	t.Parallel()
	repository := &memoryRepository{}
	service := NewService(repository)
	registration, err := service.Register(context.Background(), RegisterInput{Email: "DEV@EXAMPLE.COM", DisplayName: "Dev", Password: "a-secure-password", OrganizationName: "Acme", OrganizationSlug: "acme"})
	if err != nil {
		t.Fatal(err)
	}
	if registration.User.Email != "dev@example.com" {
		t.Fatalf("email=%q", registration.User.Email)
	}
	if registration.Organization.ID == "" {
		t.Fatal("registration did not return the organization")
	}
	session, err := service.Login(context.Background(), "dev@example.com", "a-secure-password")
	if err != nil {
		t.Fatal(err)
	}
	if session.Token == "" || repository.tokenHash == session.Token {
		t.Fatal("session token was not hashed at rest")
	}
	if _, err := service.Authenticate(context.Background(), session.Token); err != nil {
		t.Fatal(err)
	}
}

func TestLoginUsesGenericUnauthorizedError(t *testing.T) {
	t.Parallel()
	service := NewService(&memoryRepository{})
	if _, err := service.Login(context.Background(), "missing@example.com", "wrong"); err != domain.ErrUnauthorized {
		t.Fatalf("error=%v", err)
	}
}

func TestSessionJSONContract(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(Session{Token: "token", ExpiresAt: time.Unix(0, 0).UTC(), User: models.User{ID: "user-id"}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"token", "expiresAt", "user"} {
		if _, ok := fields[name]; !ok {
			t.Fatalf("missing JSON field %q in %s", name, payload)
		}
	}
	for _, name := range []string{"Token", "ExpiresAt", "User"} {
		if _, ok := fields[name]; ok {
			t.Fatalf("unexpected legacy JSON field %q in %s", name, payload)
		}
	}
}
