package integrationtest

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

type TokenSpec struct {
	UserID         string
	Email          string
	Role           model.Role
	IssuedAt       time.Time
	ExpiresAt      time.Time
	SigningSecret  string
}

func (e *TestEnv) SignToken(spec TokenSpec) (string, error) {
	userID := spec.UserID
	if userID == "" {
		userID = uuid.NewString()
	}

	email := spec.Email
	if email == "" {
		email = "integration@example.com"
	}

	role := spec.Role
	if role == "" {
		role = model.RoleEmployee
	}

	issuedAt := spec.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}

	expiresAt := spec.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = issuedAt.Add(15 * time.Minute)
	}

	secret := spec.SigningSecret
	if secret == "" {
		secret = e.Config.JWTSecretKey
	}

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  string(role),
		"iat":   issuedAt.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (e *TestEnv) BearerToken(spec TokenSpec) (string, error) {
	signed, err := e.SignToken(spec)
	if err != nil {
		return "", err
	}
	return "Bearer " + signed, nil
}

func (e *TestEnv) MustBearerToken(t testing.TB, userID uuid.UUID, role model.Role) string {
	t.Helper()

	token, err := e.BearerToken(TokenSpec{
		UserID: userID.String(),
		Role:   role,
	})
	if err != nil {
		t.Fatalf("failed to build bearer token: %v", err)
	}
	return token
}

func (e *TestEnv) MustExpiredBearerToken(t testing.TB, userID uuid.UUID, role model.Role) string {
	t.Helper()

	now := time.Now()
	token, err := e.BearerToken(TokenSpec{
		UserID:    userID.String(),
		Role:      role,
		IssuedAt:  now.Add(-2 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to build expired bearer token: %v", err)
	}
	return token
}

func (e *TestEnv) MustInvalidSignatureBearerToken(t testing.TB, userID uuid.UUID, role model.Role) string {
	t.Helper()

	token, err := e.BearerToken(TokenSpec{
		UserID:        userID.String(),
		Role:          role,
		SigningSecret: e.Config.JWTSecretKey + "-invalid",
	})
	if err != nil {
		t.Fatalf("failed to build invalid-signature bearer token: %v", err)
	}
	return token
}
