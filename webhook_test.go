package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func sendWebhookRequest(t *testing.T, payload GithubResponse, headers map[string]string) (*http.Response, map[string]any) {
	client := &http.Client{}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		t.Errorf("Error when marshal json %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/webhook", bytes.NewBuffer(jsonPayload))

	if err != nil {
		t.Errorf("Failed to create HTTP request %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	mac := hmac.New(sha256.New, []byte("helloworld"))

	mac.Write(bytes.NewBuffer(jsonPayload).Bytes())
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-HUB-SIGNATURE-256", fmt.Sprintf("sha256=%v", computedSignature))

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

func TestCreateNewServer(t *testing.T) {
	t.Setenv("PORT", "8080")

	err := NewServer()
	defer Server.Shutdown(t.Context())

	assert.NoError(t, err, "Create new server should not return an error")
}

func TestShutdownServer(t *testing.T) {
	t.Setenv("PORT", "8080")

	err := NewServer()

	if err != nil {
		t.Errorf("Creating server error %v", err)
	}

	time.Sleep(1 * time.Second)

	err = ShutdownServer()

	assert.NoError(t, err, "Shutdown server should not return an error")
}

func TestWebhookSuccess(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "helloworld")

	err := NewServer()

	if err != nil {
		t.Errorf("Creating server error %v", err)
	}

	defer Server.Shutdown(t.Context())

	githubResponse := GithubResponse{
		Repository: RepositoryStruct{
			HtmlUrl: "https://github.com/khouwdevin/gitomatically"},
		Ref: "refs/heads/master",
	}

	headers := map[string]string{
		"X-Github-Event": "push",
	}

	res, jsonResponse := sendWebhookRequest(t, githubResponse, headers)

	message := jsonResponse["message"].(string)

	assert.Equal(t, "Webhook receive", message, "Webhook response should return Webhook receive")
	assert.Equal(t, http.StatusOK, res.StatusCode, "Webhook response code should return 200")
}
