package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GithubAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		signatureHeader := c.GetHeader("X-Hub-Signature-256")

		if signatureHeader == "" {
			slog.Debug("MIDDLEWARE Signature is not found")

			c.JSON(http.StatusUnauthorized, gin.H{"message": "X-Hub-Signature-256 is not found!"})
			c.Abort()

			return
		}

		expectedSignature := signatureHeader[7:]

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error(fmt.Sprintf("MIDDLEWARE Error reading body %v", err))

			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			c.Abort()

			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))

		mac.Write(bodyBytes)
		computedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(computedSignature), []byte(expectedSignature)) {
			slog.Debug("MIDDLEWARE Signature is not match")

			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized!"})
			c.Abort()

			return
		}

		c.Next()
	}
}
