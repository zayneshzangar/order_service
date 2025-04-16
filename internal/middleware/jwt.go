package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type key string

var secretKey = os.Getenv("JWT_SECRET_KEY")

const (
	UserIDKey key = "user_id"
	RoleKey   key = "role"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		// 1. Сначала пробуем из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// 2. Если нет — пробуем из cookie
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}
			tokenStr = cookie.Value
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			fmt.Println("❌ JWT error:", err)
			fmt.Println("❌ Token string:", tokenStr)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user_id", http.StatusUnauthorized)
			return
		}
		userID := int64(userIDFloat)

		role, _ := claims["role"].(string)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

func ParseToken(tokenString string) (int64, string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return 0, "", err
	}

	// Приводим user_id к int64
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", errors.New("invalid user_id")
	}
	userID := int64(userIDFloat)

	// Получаем роль
	role, ok := claims["role"].(string)
	if !ok {
		return 0, "", errors.New("invalid role")
	}

	return userID, role, nil
}
