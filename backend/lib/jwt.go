	package lib

	import (
		"errors"
		"os"
		"time"

		"github.com/golang-jwt/jwt/v5"
	)

	// generateToken creates a JWT token with user_id, issued at, expiration, and not before claims.
	func generateToken(userID string, secret string, expiry time.Duration) (string, error) {
		now := time.Now()

		claims := jwt.MapClaims{
			"user_id": userID,
			"iat":     now.Unix(),               // Issued At
			"nbf":     now.Unix(),               // Not Before
			"exp":     now.Add(expiry).Unix(),  // Expiration Time
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signedToken, err := token.SignedString([]byte(secret))
		if err != nil {
			return "", err
		}

		return signedToken, nil
	}

	// GenerateAccessToken generates an access token valid for 15 minutes.
	func GenerateAccessToken(userID string) (string, error) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return "", errors.New("JWT_SECRET environment variable is not set")
		}
		return generateToken(userID, secret, 24*time.Hour)
	}

	// GenerateRefreshToken generates a refresh token valid for 7 days.
	func GenerateRefreshToken(userID string) (string, error) {
		secret := os.Getenv("JWT_REFRESH_SECRET")
		if secret == "" {
			return "", errors.New("JWT_REFRESH_SECRET environment variable is not set")
		}
		return generateToken(userID, secret, 7*24*time.Hour)
	}

	// VerifyToken parses and validates a token string using the given secret and returns the claims.
	func VerifyToken(tokenStr, secret string) (jwt.MapClaims, error) {
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil {
			return nil, err
		}

		if !token.Valid {
			return nil, errors.New("token is invalid")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("failed to parse token claims")
		}

		return claims, nil
	}
