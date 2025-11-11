package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"fulcrum/ports"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleRegister_Ok(t *testing.T) {
	// Arrange
	randomData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	api, fakes := createTestAPI(t, randomData)
	username := "testuser"
	expectedSecret := base64.URLEncoding.EncodeToString(randomData)
	expectedPlayerID := 123
	currentPhaseID := 9
	fakes.store.RegisterPlayerFunc = func(ctx context.Context, playerName, secret string) (int, error) {
		assert.Equal(t, username, playerName)
		assert.Equal(t, expectedSecret, secret)
		return expectedPlayerID, nil
	}
	fakes.rules.CurrentPhase = currentPhaseID

	// Act
	rec := CallAPI(t, "/register", RegisterRequest{Username: username}, api.handleRegister)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp RegisterResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, expectedPlayerID, resp.PlayerID)
	assert.Equal(t, username, resp.Username)
	assert.Equal(t, expectedSecret, resp.Secret)
	assert.Equal(t, currentPhaseID, resp.PhaseID)
}

func TestHandleRegister_EmptyUsername(t *testing.T) {
	// Arrange
	api, _ := createTestAPI(t, []byte{})

	// Act
	rec := CallAPI(t, "/register", RegisterRequest{Username: ""}, api.handleRegister)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Username is required", resp["error"])
}

func TestHandleRegister_InvalidJSON(t *testing.T) {
	// Arrange
	api, _ := createTestAPI(t, []byte{})

	// Act
	rec := CallAPI(t, "/register", "invalid json", api.handleRegister)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Invalid request body", resp["error"])
}

func TestHandleRegister_StoreFail(t *testing.T) {
	// Arrange
	randomData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	api, fakes := createTestAPI(t, randomData)
	fakes.store.RegisterPlayerFunc = func(ctx context.Context, playerName, secret string) (int, error) {
		return 0, fmt.Errorf("failed to register player")
	}
	fakes.store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
		return 1, nil
	}

	// Act
	rec := CallAPI(t, "/register", RegisterRequest{Username: "testuser"}, api.handleRegister)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Failed to register player", resp["error"])
}

func TestHandleRegister_GamePhaseIsNotBegin(t *testing.T) {
	// ARRANGE
	randomData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	api, fakes := createTestAPI(t, randomData)
	fakes.rules.RegistrationError = ports.ErrorIncorrectInput

	// ACT
	rec := CallAPI(t, "/register", RegisterRequest{Username: "testuser"}, api.handleRegister)

	// ASSERT
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
}
