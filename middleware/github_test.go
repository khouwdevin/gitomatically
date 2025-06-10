package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func initializeServer(t *testing.T) *http.Server {
	router := gin.Default()

	router.POST("/webhook", GithubAuthorization(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Webhook receive"})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Gin server error %v", err)
		}
	}()

	return server
}

func sendRequest(t *testing.T, payload map[string]any, githubWebhookSecret string, skipHeader bool) (*http.Response, map[string]any) {
	client := &http.Client{}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		t.Errorf("Error when marshal json %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/webhook", bytes.NewBuffer(jsonPayload))

	if err != nil {
		t.Errorf("Failed to create HTTP request %v", err)
	}

	mac := hmac.New(sha256.New, []byte(githubWebhookSecret))

	mac.Write(bytes.NewBuffer(jsonPayload).Bytes())
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	if !skipHeader {
		req.Header.Set("X-HUB-SIGNATURE-256", fmt.Sprintf("sha256=%v", computedSignature))
	}

	res, err := client.Do(req)

	if err != nil {
		t.Errorf("Failed to send request %v", err)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("Failed to read API response %v", err)
	}

	var jsonResponse map[string]any

	err = json.Unmarshal(bodyBytes, &jsonResponse)

	if err != nil {
		t.Errorf("Failed to unmarshall response %v", err)
	}

	return res, jsonResponse
}

func TestGithubMiddlewareSuccess(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEHBOOK_SECRET")
	})

	t.Setenv("GITHUB_WEBHOOK_SECRET", "helloworld")

	Server := initializeServer(t)
	defer Server.Shutdown(t.Context())

	_, jsonResponse := sendRequest(t, map[string]any{"message": "webhook testing"}, "helloworld", false)
	message := jsonResponse["message"].(string)

	assert.Equal(t, "Webhook receive", message, "API response message should return Webhook receive")
}

func TestGithubMiddlwareUnauthorized(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEHBOOK_SECRET")
	})

	t.Setenv("GITHUB_WEBHOOK_SECRET", "helloworld")

	Server := initializeServer(t)
	defer Server.Shutdown(t.Context())

	res, jsonResponse := sendRequest(t, map[string]any{"message": "webhook testing"}, "worldhello", false)
	message := jsonResponse["message"].(string)

	assert.Equal(t, "Unauthorized!", message, "API response message should return unauthorized")
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "API status should return 401 (unauthorized)")
}

func TestGithubMiddlwareSecretNotFound(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEHBOOK_SECRET")
	})

	t.Setenv("GITHUB_WEBHOOK_SECRET", "helloworld")

	Server := initializeServer(t)
	defer Server.Shutdown(t.Context())

	res, jsonResponse := sendRequest(t, map[string]any{"message": "webhook testing"}, "worldhello", true)
	message := jsonResponse["message"].(string)

	assert.Equal(t, "X-Hub-Signature-256 is not found!", message, "API response message should return X-Hub-Signature-256 is not found!")
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "API status should return 401 (unauthorized)")
}
