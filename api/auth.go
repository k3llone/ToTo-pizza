package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
	"toto-pizza/database"

	"github.com/gin-gonic/gin"
)

func generateRandomString() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func CreateSession(userId uint) (string, error) {
	tokenString, err := generateRandomString()

	if err != nil {
		return "", err
	}

	database.DB.Create(&database.Session{SessionId: tokenString, UserId: userId})

	return tokenString, nil
}

func CheckSession(sessionId string) bool {
	var session database.Session
	a := database.DB.First(&session, "session_id = ?", sessionId)

	if a.Error != nil {
		return false
	}

	endTime := session.CreatedAt
	endTime.Add(24 * time.Hour)

	return !time.Now().Before(endTime)
}

func Login(phone string, password string) error {
	var currentUser database.User

	database.DB.First(&currentUser, "phone = ?", phone)

	if GetHash(password) != currentUser.PasswordHash {
		return errors.New("password")
	}

	_, err := CreateSession(currentUser.ID)

	if err != nil {
		return err
	}

	return nil
}

func GetHash(str string) string {
	bytesHash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(bytesHash[:])
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		if !CheckSession(token) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Session has expired"})
			return
		} else {
			c.Next()
		}

		return
	}
}

func GetUserByToken(token string) database.User {
	var User database.User
	var Session database.Session

	database.DB.First(&Session, "session_id = ?", token)
	database.DB.First(&User, "id = ?", Session.UserId)

	return User
}
