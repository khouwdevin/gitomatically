package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khouwdevin/gitomatically/env"
)

func GithubAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		signatureHeader := c.GetHeader("X-Hub-Signature-256")

		if signatureHeader == "" {
			slog.Debug("[Middleware] Signature is not found")

			c.JSON(http.StatusUnauthorized, gin.H{"Message": "X-Hub-Signature-256 is not found!"})
			c.Abort()

			return
		}

		expectedSignature := signatureHeader[7:]

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(env.Env.GITHUB_WEBHOOK_SECRET))

		mac.Write(bodyBytes)
		computedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(computedSignature), []byte(expectedSignature)) {
			slog.Debug("[Middleware] Signature is not match")

			c.JSON(http.StatusUnauthorized, gin.H{"message": "StatusUnauthorized"})
			c.Abort()

			return
		}

		c.Next()
	}
}
