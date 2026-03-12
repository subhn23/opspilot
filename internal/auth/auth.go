package auth

import (
	"errors"
	"log"
	"opspilot/internal/audit"
	"opspilot/internal/config"
	"opspilot/internal/models"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

// JWTClaims represents the custom claims for the JWT
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT for a user
func GenerateToken(userID, roleID uuid.UUID, secret string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT string
func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateTOTPSecret creates a new MFA secret for a user's enrollment
func GenerateTOTPSecret(email string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpsPilot",
		AccountName: email,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// VerifyTOTP validates the 6-digit MFA code against the user's secret
func VerifyTOTP(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

// AuthMiddleware verifies the JWT in the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			audit.LogAction(config.DB, uuid.Nil, "AUTH_FAILURE", "Invalid Format", c.ClientIP(), authHeader)
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization format"})
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Println("JWT_SECRET not set in environment")
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}

		claims, err := ValidateToken(parts[1], secret)
		if err != nil {
			audit.LogAction(config.DB, uuid.Nil, "AUTH_FAILURE", "Invalid Token", c.ClientIP(), parts[1])
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Validate user existence (Session Validation)
		var user models.User
		if err := config.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
			audit.LogAction(config.DB, uuid.Nil, "AUTH_FAILURE", "User Not Found", c.ClientIP(), claims.UserID.String())
			c.AbortWithStatusJSON(401, gin.H{"error": "User no longer exists"})
			return
		}

		// Set identity in context
		c.Set("user_id", claims.UserID)
		c.Set("role_id", claims.RoleID)

		c.Next()
	}
}
