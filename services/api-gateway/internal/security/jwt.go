package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID string   `json:"sub"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey []byte
	issuer    string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    "ekyc-api-gateway",
	}
}

// GenerateToken generates a new JWT token for a user
func (j *JWTManager) GenerateToken(userID string, roles []string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// HasRole checks if the user has a specific role
func (j *JWTManager) HasRole(claims *Claims, requiredRole string) bool {
	for _, role := range claims.Roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the required roles
func (j *JWTManager) HasAnyRole(claims *Claims, requiredRoles ...string) bool {
	for _, requiredRole := range requiredRoles {
		if j.HasRole(claims, requiredRole) {
			return true
		}
	}
	return false
}

// ExtractUserIDFromToken extracts user ID from token without full validation
func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (string, error) {
	// Parse without validation for quick extraction
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}

// TokenExpiration returns the default token expiration time
func (j *JWTManager) TokenExpiration() time.Duration {
	return 24 * time.Hour // 24 hours
}
