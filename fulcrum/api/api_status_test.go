package api

import (
	"context"
	"encoding/json"
	"fulcrum/ports"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleStatus_NominalCase(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return secret == "valid-secret", nil
	}
	fakes.store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
		return 13, nil
	}
	fakes.rules.RelevantAssets = []string{"SH", "SF"}
	fakes.rules.Scores = []int{1, 2}

	// Act
	resp := CallAPI(t, "/status", StatusRequest{
		PlayerID: 1,
		Secret:   "valid-secret",
	}, api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	var response StatusResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 13, response.CurrentPhase)
	assert.NotContains(t, response.SelectedAssets, "RB")
	assert.Contains(t, response.SelectedAssets, "SH")
	assert.NotContains(t, response.SelectedAssets, "RD")
	assert.Contains(t, response.SelectedAssets, "SF")
	assert.Equal(t, 1, response.Scores[0])
	assert.Equal(t, 2, response.Scores[1])
}

func TestHandleStatus_RejectInvalidPlayerSecret(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return secret == "valid-secret", nil
	}

	// Act
	resp := CallAPI(t, "/status", StatusRequest{
		PlayerID: 1,
		Secret:   "invalid-secret",
	}, api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	var errorResponse map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Invalid player credentials")
}

func TestHandleStatus_ErrorVerifyingPlayerSecret(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return false, assert.AnError
	}

	// Act
	resp := CallAPI(t, "/status", StatusRequest{
		PlayerID: 1,
		Secret:   "secret",
	}, api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Failed to verify player identity")
}

func TestHandleStatus_ErrorGettingCurrentPhase(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return secret == "valid-secret", nil
	}
	fakes.store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
		return 0, assert.AnError
	}

	// Act
	resp := CallAPI(t, "/status", StatusRequest{
		PlayerID: 1,
		Secret:   "valid-secret",
	}, api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Failed to get current phase")
}

func TestHandleStatus_ErrorGettingPlayerAssets(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return secret == "valid-secret", nil
	}
	fakes.store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
		return 3, nil
	}
	fakes.rules.StatusError = ports.ErrorSystemFailure

	// Act
	resp := CallAPI(t, "/status", StatusRequest{
		PlayerID: 1,
		Secret:   "valid-secret",
	}, api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
}

func TestHandleStatus_InvalidRequestBody(t *testing.T) {
	// Arrange
	randomData := []byte("randomsecretdata")
	api, _ := createTestAPI(t, randomData)

	// Act
	rec := CallAPI(t, "/status", "{invalid json}", api.handleStatus)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Invalid request body")
}
