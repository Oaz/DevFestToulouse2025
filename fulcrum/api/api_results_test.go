package api

import (
	"encoding/json"
	"fulcrum/ports"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleResults_ValidRequest(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{ID: 1, Assets: map[string]int{"asset1": 2, "asset2": 1}, Score: 100},
		{ID: 2, Assets: map[string]int{"asset3": 3}, Score: 150},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 1, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_InvalidAdminPassword(t *testing.T) {
	api, _ := createTestAPI(t, []byte("random"))

	request := ResultsRequest{
		AdminPassword: "wrongpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), "Invalid admin password")
}

func TestHandleResults_EmptyAdminPassword(t *testing.T) {
	api, _ := createTestAPI(t, []byte("random"))

	request := ResultsRequest{
		AdminPassword: "",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), "Invalid admin password")
}

func TestHandleResults_InvalidRequestBody(t *testing.T) {
	api, _ := createTestAPI(t, []byte("random"))

	response := CallAPI(t, "/results", "invalid json", api.handleResults)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), "Invalid request body")
}

func TestHandleResults_RulesSystemFailure(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	fakes.rules.StatusError = ports.ErrorSystemFailure
	fakes.rules.PlayerResults = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), ports.ErrorSystemFailure.Error())
}

func TestHandleResults_RulesIncorrectInput(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	fakes.rules.StatusError = ports.ErrorIncorrectInput
	fakes.rules.PlayerResults = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), ports.ErrorIncorrectInput.Error())
}

func TestHandleResults_UnexpectedRulesError(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	fakes.rules.StatusError = assert.AnError
	fakes.rules.PlayerResults = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assertErrorResponse(t, response.Body.Bytes(), "Unexpected error")
}

func TestHandleResults_EmptyResults(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	fakes.rules.PlayerResults = []ports.PlayerResult{}
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       2,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 2, actualResponse.PhaseID)
	assert.Equal(t, []ports.PlayerResult{}, actualResponse.Results)
}

func TestHandleResults_NilResults(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	fakes.rules.PlayerResults = nil
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       3,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 3, actualResponse.PhaseID)
	assert.Nil(t, actualResponse.Results)
}

func TestHandleResults_ZeroPhaseID(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{ID: 1, Assets: map[string]int{}, Score: 0},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       0,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 0, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_NegativePhaseID(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{ID: 1, Assets: map[string]int{}, Score: 0},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       -1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, -1, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_LargePhaseID(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{ID: 1000, Assets: map[string]int{"asset1000": 5}, Score: 9999},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       999999,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 999999, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_LargePlayerResults(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	// Create a large set of player results
	expectedResults := make([]ports.PlayerResult, 100)
	for i := 0; i < 100; i++ {
		expectedResults[i] = ports.PlayerResult{
			ID:     i + 1,
			Assets: map[string]int{"asset1": i, "asset2": i * 2},
			Score:  i * 10,
		}
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 1, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_SinglePlayerResult(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{
			ID:     42,
			Assets: map[string]int{"singleAsset": 1},
			Score:  500,
		},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       1,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 1, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

func TestHandleResults_PlayerResultsWithZeroAssetCounts(t *testing.T) {
	api, fakes := createTestAPI(t, []byte("random"))

	expectedResults := []ports.PlayerResult{
		{
			ID:     1,
			Assets: map[string]int{"asset1": 0, "asset2": 5, "asset3": 0},
			Score:  100,
		},
		{
			ID:     2,
			Assets: map[string]int{"asset1": 10, "asset2": 0},
			Score:  300,
		},
	}

	fakes.rules.PlayerResults = expectedResults
	fakes.rules.StatusError = nil

	request := ResultsRequest{
		AdminPassword: "verysecretpassword",
		PhaseID:       2,
	}

	response := CallAPI(t, "/results", request, api.handleResults)

	assert.Equal(t, http.StatusOK, response.Code)

	var actualResponse ResultsResponse
	err := parseJSONResponse(t, response.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)

	assert.Equal(t, 2, actualResponse.PhaseID)
	assert.Equal(t, expectedResults, actualResponse.Results)
}

// Helper functions for testing

func parseJSONResponse(t *testing.T, body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func assertErrorResponse(t *testing.T, body []byte, expectedMessage string) {
	var errorResponse map[string]string
	err := json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, errorResponse["error"])
}
