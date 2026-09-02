package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/jackc/pgx/v5"
)

type AuthRepository struct{ db *DB }

func NewAuthRepository(db *DB) *AuthRepository { return &AuthRepository{db: db} }

func (r *AuthRepository) Register(ctx context.Context, user models.User, organization models.Organization) error {
	return r.db.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO users (id,email,display_name,password_hash,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$5)`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt); err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO organizations (id,name,slug,created_at) VALUES ($1,$2,$3,$4)`, organization.ID, organization.Name, organization.Slug, organization.CreatedAt); err != nil {
			return fmt.Errorf("create organization: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO memberships (organization_id,user_id,role) VALUES ($1,$2,'owner')`, organization.ID, user.ID); err != nil {
			return fmt.Errorf("create owner membership: %w", err)
		}
		return nil
	})
}

func (r *AuthRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$5)`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.Pool.QueryRow(ctx, `SELECT id,email,display_name,password_hash,created_at,updated_at FROM users WHERE email=$1`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, domain.ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) UserByID(ctx context.Context, id models.ID) (models.User, error) {
	var user models.User
	err := r.db.Pool.QueryRow(ctx, `SELECT id,email,display_name,password_hash,created_at,updated_at FROM users WHERE id=$1`, id).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, domain.ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) OrganizationsByUser(ctx context.Context, userID models.ID) ([]models.OrganizationMembership, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT o.id,o.name,o.slug,o.created_at,m.role FROM memberships m JOIN organizations o ON o.id=m.organization_id WHERE m.user_id=$1 ORDER BY o.name,o.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}
	defer rows.Close()
	items := []models.OrganizationMembership{}
	for rows.Next() {
		var item models.OrganizationMembership
		if err := rows.Scan(&item.Organization.ID, &item.Organization.Name, &item.Organization.Slug, &item.Organization.CreatedAt, &item.Role); err != nil {
			return nil, fmt.Errorf("scan user organization: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AuthRepository) CreateSession(ctx context.Context, id, tokenHash string, userID models.ID, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO sessions (id, token_hash, user_id, expires_at) VALUES ($1,$2,$3,$4)`, id, tokenHash, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *AuthRepository) UserBySession(ctx context.Context, tokenHash string) (models.User, error) {
	var user models.User
	err := r.db.Pool.QueryRow(ctx, `SELECT u.id,u.email,u.display_name,u.password_hash,u.created_at,u.updated_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>now()`, tokenHash).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get session: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
