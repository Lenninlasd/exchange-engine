package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// HasScope checks whether our claims have a specific scope.
func (c CustomClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

func GetTokenFromHeader(headers http.Header) string {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 {
		return ""
	}

	if strings.ToLower(tokenParts[0]) != "bearer" {
		return ""
	}

	return tokenParts[1]
}

func ValidateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Your JWT secret key
		secret := []byte(os.Getenv("JWT_SECRET"))

		// Sample JWT token to validate (replace with your actual token)
		tokenString := GetTokenFromHeader(r.Header)
		if tokenString == "" {
			http.Error(w, "Missing JWT token", http.StatusUnauthorized)
			return
		}
		// Parse the token using the provided secret
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				http.Error(w, "Invalid signature", http.StatusInternalServerError)
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				http.Error(w, "Token is expired or not yet valid", http.StatusUnauthorized)
			} else {
				fmt.Println("Invalid JWT token", err)
			}
		}

		if token.Valid {
			fmt.Println("Token is valid")
			next.ServeHTTP(w, r)
		}
		return
	})
}
