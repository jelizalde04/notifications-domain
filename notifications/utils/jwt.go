package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ParseJWT decodifica un token JWT y devuelve los claims
func ParseJWT(tokenString string) (map[string]interface{}, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	// Debug: verificar que el secret existe
	if jwtSecret == "" {
		log.Println("ERROR: JWT_SECRET environment variable is not set")
		return nil, errors.New("JWT_SECRET not configured")
	}

	log.Printf("JWT_SECRET loaded (length: %d)", len(jwtSecret))
	log.Printf("Token to parse: %s", tokenString[:50]+"...") // Solo primeros 50 chars por seguridad

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validar algoritmo
		log.Printf("JWT Algorithm: %v", token.Method)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Printf("JWT Parse error: %v", err)
		return nil, fmt.Errorf("token parse failed: %v", err)
	}

	if !token.Valid {
		log.Println("JWT token is not valid")
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("JWT claims are not valid")
		return nil, errors.New("invalid token claims")
	}

	// Verificar expiración
	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := int64(exp)
		currentTime := time.Now().Unix()
		log.Printf("Token expiration: %d, Current time: %d", expirationTime, currentTime)

		if expirationTime < currentTime {
			log.Printf("Token expired: exp=%d, now=%d", expirationTime, currentTime)
			return nil, errors.New("token expired")
		}
	} else {
		log.Println("No expiration claim found in token")
	}

	log.Printf("JWT parsed successfully, claims: %v", claims)
	return claims, nil
}
