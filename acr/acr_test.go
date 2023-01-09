package acr

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	testSigningKey []byte = []byte("AllYourBase")
	testTenantID   string = "test_tenant_id"
)

type TestClaims struct {
	jwt.RegisteredClaims
	TenantID string `json:"tid,omitempty"`
}

func generateJWT(claims TestClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(testSigningKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func TestGetTenantIDWithTenantID(t *testing.T) {
	claims := TestClaims{
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
			Subject:   "somebody",
			ID:        "1",
			Audience:  []string{"somebody_else"},
		},
		testTenantID,
	}

	jwt, err := generateJWT(claims)
	if err != nil {
		t.Fatalf("unable to generate a jwt: %s", err)
	}

	tenantID, err := getTenantID(jwt)
	if err != nil {
		t.Fatalf("unable to fetch tenant id: %s", err)
	}

	if tenantID != testTenantID {
		t.Fatalf("fetched tenant id does not match desired tenant id: %s != %s", tenantID, testTenantID)
	}
}

func TestGetTenantIDWithoutTenantID(t *testing.T) {
	claims := TestClaims{
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
			Subject:   "somebody",
			ID:        "1",
			Audience:  []string{"somebody_else"},
		},
		"",
	}

	jwt, err := generateJWT(claims)
	if err != nil {
		t.Fatalf("unable to generate a jwt: %s", err)
	}

	_, err = getTenantID(jwt)
	if err == nil || (err != nil && err != ErrNoTenantIDClaim) {
		t.Fatalf("expected error 'ErrNoTenantIDClaim', got error: %s", err)
	}
}
