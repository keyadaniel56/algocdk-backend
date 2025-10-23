// package middleware

// import (
// 	"Api/utils"
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// )

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		authHeader := ctx.GetHeader("Authorization")
// 		if authHeader == "" {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
// 			ctx.Abort()
// 			return
// 		}

// 		// support "Bearer <token>" or raw token
// 		tokenString := strings.TrimSpace(authHeader)
// 		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
// 			tokenString = strings.TrimSpace(tokenString[7:])
// 		}

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			return utils.JwtKey, nil
// 		})
// 		if err != nil || !token.Valid {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
// 			ctx.Abort()
// 			return
// 		}
// 		claims := token.Claims.(jwt.MapClaims)
// 		ctx.Set("user_id", uint(claims["user_id"].(float64)))
// 		ctx.Set("email", claims["email"].(string))

// 		ctx.Next()
// 	}
// }

package middleware

import (
	"Api/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try getting token from header first
		tokenString := ctx.GetHeader("Authorization")

		// If header is empty, try query param (for WebSocket connections)
		if tokenString == "" {
			tokenString = ctx.Query("token")
		}

		// If token still missing, block request
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			ctx.Abort()
			return
		}

		// Handle “Bearer <token>” format
		tokenString = strings.TrimSpace(tokenString)
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			tokenString = strings.TrimSpace(tokenString[7:])
		}

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return utils.JwtKey, nil
		})
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			ctx.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx.Set("user_id", uint(claims["user_id"].(float64)))
		ctx.Set("email", claims["email"].(string))
		
		ctx.Next()
	}
}
