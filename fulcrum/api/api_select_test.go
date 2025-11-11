package api

import (
	"context"
	"encoding/json"
	"fulcrum/ports"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var FakeErr = &errFake{}

type errFake struct{}

func (e *errFake) Error() string {
	return "fake error"
}

func TestHandleSelect_ValidSelection(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp SelectResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "Assets selected successfully", resp.Message)
}

func TestHandleSelect_InvalidSelectionRequirements(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorIncorrectInput

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"wrong-asset"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_ErrorGettingPlayerAssets(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorSystemFailure

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_ErrorUpdatingPlayerAssets(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorSystemFailure

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_InvalidSecret(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return false, nil
	}

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "wrong-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, resp["error"], "Invalid player credentials")
}

func TestHandleSelect_WrongPhaseID(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorIncorrectInput

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  1, // Wrong phase ID (not matching current)
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_ErrorGettingCurrentPhase(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorSystemFailure

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_MissingPhaseConfiguration(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return true, nil
	}
	fakes.rules.AssetSelectionError = ports.ErrorSystemFailure

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  999,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}

func TestHandleSelect_InvalidRequestBody(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, _ := createTestAPI(t, randomData)

	// Act
	rec := CallAPI(t, "/select", "invalid json", api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, resp["error"], "Invalid request body")
}

func TestHandleSelect_ErrorVerifyingPlayerSecret(t *testing.T) {
	// Arrange
	randomData := make([]byte, 16)
	api, fakes := createTestAPI(t, randomData)
	fakes.store.VerifyPlayerSecretFunc = func(ctx context.Context, playerID int, secret string) (bool, error) {
		return false, FakeErr
	}

	// Act
	rec := CallAPI(t, "/select", SelectRequest{
		PlayerID: 1,
		Secret:   "test-secret",
		PhaseID:  2,
		AssetIDs: []string{"asset1"},
	}, api.handleSelect)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, resp["error"], "Failed to verify player identity")

}
