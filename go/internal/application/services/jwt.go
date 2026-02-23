package services

import (
	"fmt"
	"time"

	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret    []byte
	expiresIn time.Duration
}

// Claims is now defined in internal/types/services.go

func NewJWTService(secret string, expiresIn time.Duration) *JWTService {
	return &JWTService{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func (s *JWTService) GenerateToken(userID, email, name string, isTeacher bool) (string, error) {
	now := time.Now()
	claims := types.Claims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		IsTeacher: isTeacher,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "your-app-name", // Replace with your app name
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*types.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateTokenAllowExpired validates token structure but allows expired tokens
// Useful for refresh token endpoint
func (s *JWTService) ValidateTokenAllowExpired(tokenString string) (*types.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Manually check if token is not too old (e.g., allow tokens up to 1 hour past expiry for refresh)
	if claims.ExpiresAt != nil && time.Since(claims.ExpiresAt.Time) > time.Hour {
		return nil, fmt.Errorf("token too old for refresh")
	}

	return claims, nil
}

// GetTokenExpiry returns the expiration time of a token without full validation
func (s *JWTService) GetTokenExpiry(tokenString string) (*time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*types.Claims); ok && claims.ExpiresAt != nil {
		expiry := claims.ExpiresAt.Time
		return &expiry, nil
	}

	return nil, fmt.Errorf("no expiry time found")
}

// IsTokenExpired checks if a token is expired without full validation
func (s *JWTService) IsTokenExpired(tokenString string) bool {
	expiry, err := s.GetTokenExpiry(tokenString)
	if err != nil {
		return true // Treat parsing errors as expired
	}
	return time.Now().After(*expiry)
}
