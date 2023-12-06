package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

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
			next.ServeHTTP(w, r)
		}
		return
	})
}

type MyCustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// given a token string parse the jwt and extract the clains
func GetClainsFromToken(header http.Header) (*MyCustomClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	tokenString := GetTokenFromHeader(header)

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	claims, ok := token.Claims.(*MyCustomClaims)

	if ok && token.Valid {
		fmt.Printf("%v %v", claims.Email, claims.RegisteredClaims.Issuer)
		return claims, nil
	}

	fmt.Println(err)
	return claims, err
}
