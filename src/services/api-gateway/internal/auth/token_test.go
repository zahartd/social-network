package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeys(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate RSA key for test")
	return privKey
}

func setupTestPublicKey(t *testing.T, privKey *rsa.PrivateKey) {
	t.Helper()
	rsaPublicKey = &privKey.PublicKey
	t.Cleanup(func() {
		rsaPublicKey = nil
	})
}

func createTestToken(t *testing.T, privKey *rsa.PrivateKey, claims jwt.Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, err := token.SignedString(privKey)
	require.NoError(t, err, "Failed to sign test token")
	return tokenStr
}

func TestValidateJWT_ValidToken(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	testUserID := "user-12345"
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(1 * time.Hour)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testUserID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt.Add(-1 * time.Minute)),
			Issuer:    "test-issuer",
			Audience:  jwt.ClaimStrings{"test-audience"},
		},
	}

	tokenStr := createTestToken(t, privKey, claims)
	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.NoError(t, err, "validateJWT should not return error for a valid token")
	require.NotNil(t, parsedClaims, "validateJWT should return non-nil claims for a valid token")

	assert.Equal(t, testUserID, parsedClaims.UserID, "Parsed UserID should match the original Subject claim")
	assert.Equal(t, testUserID, parsedClaims.Subject)
	assert.Equal(t, "test-issuer", parsedClaims.Issuer)
	assert.ElementsMatch(t, jwt.ClaimStrings{"test-audience"}, parsedClaims.Audience)
	assert.WithinDuration(t, expiresAt, parsedClaims.ExpiresAt.Time, time.Second, "Expiration time mismatch")
	assert.WithinDuration(t, issuedAt, parsedClaims.IssuedAt.Time, time.Second, "IssuedAt time mismatch")
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-expired",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	tokenStr := createTestToken(t, privKey, claims)
	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err, "validateJWT should return an error for an expired token")
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired), "Error should be jwt.ErrTokenExpired")
	assert.Nil(t, parsedClaims, "Claims should be nil for an expired token")
}

func TestValidateJWT_NotValidYet(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-nbf",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}

	tokenStr := createTestToken(t, privKey, claims)
	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err, "validateJWT should return an error for a token used before nbf")
	assert.True(t, errors.Is(err, jwt.ErrTokenNotValidYet), "Error should be jwt.ErrTokenNotValidYet")
	assert.Nil(t, parsedClaims, "Claims should be nil for a token used before nbf")
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	privKey1 := generateTestKeys(t)
	privKey2 := generateTestKeys(t)
	setupTestPublicKey(t, privKey2)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-sig",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenStr := createTestToken(t, privKey1, claims)
	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err, "validateJWT should return an error for invalid signature")
	assert.True(t, errors.Is(err, jwt.ErrTokenSignatureInvalid), "Error should be jwt.ErrTokenSignatureInvalid")
	assert.Nil(t, parsedClaims, "Claims should be nil for invalid signature")
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	malformedToken := "Bearer this.is.not.a.jwt"

	parsedClaims, err := validateJWT(malformedToken)

	assert.Error(t, err, "validateJWT should return an error for a malformed token")
	assert.ErrorContains(t, err, "token validation failed")
	assert.True(t, errors.Is(err, jwt.ErrTokenMalformed) ||
		errors.Is(err, jwt.ErrTokenUnverifiable) ||
		errors.Is(err, jwt.ErrTokenInvalidClaims),
		"Expected a JWT parsing error")
	assert.Nil(t, parsedClaims, "Claims should be nil for a malformed token")
}

func TestValidateJWT_NoSubject(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	mapClaims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)).Unix(),
		"iat": jwt.NewNumericDate(time.Now()).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)
	tokenStr, err := token.SignedString(privKey)
	require.NoError(t, err)

	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err, "validateJWT should return an error when Subject is missing")
	require.NotNil(t, err)
	assert.EqualError(t, err, "missing or empty Subject claim in token", "Error message mismatch")
	assert.Nil(t, parsedClaims, "Claims should be nil when Subject is missing")
}

func TestValidateJWT_WrongSigningMethod(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-wrong-alg",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("some-secret"))
	require.NoError(t, err)

	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err, "validateJWT should return an error for wrong signing method")
	require.NotNil(t, err)
	assert.ErrorContains(t, err, "unexpected signing method: HS256")
	assert.Nil(t, parsedClaims, "Claims should be nil for wrong signing method")
}

func TestValidateJWT_NoBearerPrefix(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-no-bearer"},
	}
	tokenStr := createTestToken(t, privKey, claims)
	parsedClaims, err := validateJWT(tokenStr)

	assert.NoError(t, err)
	require.NotNil(t, parsedClaims)
	assert.Equal(t, "user-no-bearer", parsedClaims.UserID)
}

func TestValidateJWT_EmptyTokenString(t *testing.T) {
	privKey := generateTestKeys(t)
	setupTestPublicKey(t, privKey)

	parsedClaims, err := validateJWT("")
	assert.Error(t, err)
	require.NotNil(t, err)
	assert.EqualError(t, err, "token string is empty after trimming Bearer prefix")
	assert.Nil(t, parsedClaims)

	parsedClaims, err = validateJWT("Bearer ")

	assert.Error(t, err)
	require.NotNil(t, err)
	assert.EqualError(t, err, "token string is empty after trimming Bearer prefix")
	assert.Nil(t, parsedClaims)
}

func TestValidateJWT_KeyNotLoaded(t *testing.T) {
	privKey := generateTestKeys(t)
	rsaPublicKey = nil

	claims := &UserClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-no-key"}}
	tokenStr := createTestToken(t, privKey, claims)
	tokenStrWithBearer := "Bearer " + tokenStr

	parsedClaims, err := validateJWT(tokenStrWithBearer)

	assert.Error(t, err)
	require.NotNil(t, err)
	assert.EqualError(t, err, "authentication service is not properly configured (missing public key)")
	assert.Nil(t, parsedClaims)
}
