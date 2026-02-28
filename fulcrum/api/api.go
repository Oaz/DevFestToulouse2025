package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fulcrum/ports"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/swaggo/http-swagger"

	// Import the docs package to register the documentation
	_ "fulcrum/docs"
)

// @title Game API
// @version 1.0
// @description This is the API server for the game application
// @termsOfService http://swagger.io/terms/
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8086
// @BasePath /

// API represents the HTTP API with injected dependencies
type API struct {
	store         ports.Store
	notifier      ports.Notifier
	rules         ports.Rules
	adminPassword string
	router        *chi.Mux
	randReader    io.Reader
}

func NewAPI(
	store ports.Store,
	notifier ports.Notifier,
	rules ports.Rules,
	adminPassword string,
	reader io.Reader,
) *API {
	api := &API{
		store:         store,
		notifier:      notifier,
		rules:         rules,
		adminPassword: adminPassword,
		router:        chi.NewRouter(),
		randReader:    reader,
	}

	api.setupRoutes()
	return api
}

func (a *API) setupRoutes() {
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer)
	a.defineCorsPolicy()
	a.router.Get("/swagger/*", httpSwagger.WrapHandler)
	a.router.Get("/ping", a.handlePing)
	a.router.Get("/ws", a.handleWebSocket)
	a.router.Post("/register", a.handleRegister)
	a.router.Post("/status", a.handleStatus)
	a.router.Post("/select", a.handleSelect)
	a.router.Post("/forward", a.handleForward)
	a.router.Post("/completion", a.handleCompletion)
	a.router.Post("/results", a.handleResults)
}

func (a *API) defineCorsPolicy() {
	a.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://devfest2025.azeau.com",
			"https://devfest2025.slides.azeau.com",
			"http://localhost:5173",
			"http://localhost:1337",
		},
		AllowCredentials: false,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		MaxAge: 3600,
	}))
}

// ServeHTTP implements the http.Handler interface
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// handlePing handles the ping request
// @Summary Ping the API
// @Description Get a pong response to verify the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func (a *API) handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
}

type ForwardRequest struct {
	AdminPassword string `json:"admin_password"`
	PhaseID       int    `json:"phase_id"`
}

// handleForward handles phase transitions
// @Summary Advance the game to a new phase
// @Description Move the game forward to the specified phase
// @Tags admin
// @Accept json
// @Produce json
// @Param request body ForwardRequest true "Forward request details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 422 {object} map[string]string "Unprocessable Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /forward [post]
func (a *API) handleForward(w http.ResponseWriter, r *http.Request) {
	var req ForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AdminPassword != a.adminPassword {
		a.ReplyWithError(w, http.StatusUnprocessableEntity, "Invalid admin password")
		return
	}

	currentPhaseID, err := a.store.GetCurrentPhase(r.Context())
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to get current phase")
		return
	}

	if req.PhaseID > currentPhaseID+1 {
		a.ReplyWithError(w, http.StatusUnprocessableEntity, "Cannot skip phases")
		return
	}

	if req.PhaseID == currentPhaseID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Already in the requested phase",
			"phase_id": currentPhaseID,
		})
		return
	}

	phaseID, err := a.store.SetPhase(r.Context(), req.PhaseID)
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to update phase")
		return
	}

	if err := a.notifier.BroadcastJson(map[string]interface{}{
		"type":     "phase_change",
		"phase_id": phaseID,
	}); err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Phase updated successfully",
		"phase_id": phaseID,
	})
}

type CompletionRequest struct {
	PhaseID int `json:"phase_id"`
}

type CompletionResponse struct {
	CompletedCount int `json:"completed_count"`
	TotalPlayers   int `json:"total_players"`
}

// handleCompletion handles requests for phase completion statistics
// @Summary Get phase completion statistics
// @Description Get the number of players who have completed a specific phase and total player count
// @Tags phases
// @Accept json
// @Produce json
// @Param request body CompletionRequest true "Phase ID to check completion for"
// @Success 200 {object} CompletionResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /completion [post]
func (a *API) handleCompletion(w http.ResponseWriter, r *http.Request) {
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	completedCount, err := a.store.GetPhaseCompletion(r.Context(), req.PhaseID)
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to get phase completion data")
		return
	}
	totalPlayers, err := a.store.GetPlayersCount(r.Context())
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to get total players count")
		return
	}
	response := CompletionResponse{
		CompletedCount: completedCount,
		TotalPlayers:   totalPlayers,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type StatusRequest struct {
	PlayerID int    `json:"player_id"`
	Secret   string `json:"secret"`
}

type StatusResponse struct {
	CurrentPhase   int      `json:"current_phase"`
	Scores         []int    `json:"scores"`
	SelectedAssets []string `json:"selected_assets"`
}

// handleStatus returns the current phase and player's selected assets
// @Summary Get player status information
// @Description Get the current game phase and player's asset selections
// @Tags players
// @Accept json
// @Produce json
// @Param request body StatusRequest true "Player credentials"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 422 {object} map[string]string "Unprocessable Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /status [post]
func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if !a.IsPlayerSecretValid(w, r, req.PlayerID, req.Secret) {
		return
	}
	currentPhaseID, err := a.store.GetCurrentPhase(r.Context())
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to get current phase")
		return
	}
	assetIDs, scores, err := a.rules.GetPlayerStatus(currentPhaseID, req.PlayerID, a.store, r.Context())
	if err != nil {
		a.ProcessError(w, err)
		return
	}
	response := StatusResponse{
		CurrentPhase:   currentPhaseID,
		Scores:         scores,
		SelectedAssets: assetIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}

type ResultsRequest struct {
	AdminPassword string `json:"admin_password"`
	PhaseID       int    `json:"phase_id"`
}

type ResultsResponse struct {
	PhaseID int                  `json:"phase_id"`
	Results []ports.PlayerResult `json:"results"`
}

// handleResults returns all players' statuses for a specific phase
// @Summary Get all players' status information for a specific phase (admin only)
// @Description Get all players' statuses for the specified phase (requires admin password)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body ResultsRequest true "Admin credentials and phase ID"
// @Success 200 {object} ResultsResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 422 {object} map[string]string "Unprocessable Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /results [post]
func (a *API) handleResults(w http.ResponseWriter, r *http.Request) {
	var req ResultsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AdminPassword != a.adminPassword {
		a.ReplyWithError(w, http.StatusUnprocessableEntity, "Invalid admin password")
		return
	}

	results, err := a.rules.GetAllPlayersStatuses(req.PhaseID, a.store, r.Context())
	if err != nil {
		a.ProcessError(w, err)
		return
	}

	response := ResultsResponse{
		PhaseID: req.PhaseID,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}

type SelectRequest struct {
	PlayerID int      `json:"player_id"`
	Secret   string   `json:"secret"`
	PhaseID  int      `json:"phase_id"`
	AssetIDs []string `json:"asset_ids"`
}

type SelectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// handleSelect handles player asset selection
// @Summary Select assets for a player
// @Description Process a player's asset selection for the current phase
// @Tags players
// @Accept json
// @Produce json
// @Param request body SelectRequest true "Player selection details"
// @Success 200 {object} SelectResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 422 {object} map[string]string "Unprocessable Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /select [post]
func (a *API) handleSelect(w http.ResponseWriter, r *http.Request) {
	var req SelectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if !a.IsPlayerSecretValid(w, r, req.PlayerID, req.Secret) {
		return
	}
	err := a.rules.AssetSelection(req.PlayerID, req.PhaseID, req.AssetIDs, a.store, r.Context())
	if err != nil {
		a.ProcessError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SelectResponse{
		Success: true,
		Message: "Assets selected successfully",
	})
}

func (a *API) IsPlayerSecretValid(w http.ResponseWriter, r *http.Request, playerID int, secret string) bool {
	isValidSecret, err := a.store.VerifyPlayerSecret(r.Context(), playerID, secret)
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to verify player identity")
		return false
	}
	if !isValidSecret {
		a.ReplyWithError(w, http.StatusUnprocessableEntity, "Invalid player credentials")
		return false
	}
	return true
}

type RegisterRequest struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	PlayerID int    `json:"player_id"`
	Username string `json:"username"`
	Secret   string `json:"secret"`
	PhaseID  int    `json:"phase_id"`
}

// handleRegister handles player registration
// @Summary Register a new player
// @Description Register a new player with a username and get back a player ID and secret
// @Tags players
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Player registration details"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 422 {object} map[string]string "Unprocessable Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.ReplyWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		a.ReplyWithError(w, http.StatusBadRequest, "Username is required")
		return
	}

	phaseId, err := a.rules.EnsureRegistrationIsAllowed(a.store, r.Context())
	if err != nil {
		a.ProcessError(w, err)
		return
	}

	secret, err := a.generateRandomSecret(16) // 16 bytes = 128 bits
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to generate secret")
		return
	}

	playerID, err := a.store.RegisterPlayer(r.Context(), req.Username, secret)
	if err != nil {
		a.ReplyWithError(w, http.StatusInternalServerError, "Failed to register player")
		return
	}

	response := RegisterResponse{
		PlayerID: playerID,
		Username: req.Username,
		Secret:   secret,
		PhaseID:  phaseId,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *API) ProcessError(w http.ResponseWriter, err error) {
	if errors.Is(err, ports.ErrorSystemFailure) {
		a.ReplyWithError(w, http.StatusInternalServerError, err.Error())
	} else if errors.Is(err, ports.ErrorIncorrectInput) {
		a.ReplyWithError(w, http.StatusUnprocessableEntity, err.Error())
	} else {
		a.ReplyWithError(w, http.StatusInternalServerError, "Unexpected error")
	}
}

func (a *API) ReplyWithError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

func (a *API) generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := a.randReader.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// handleWebSocket handles WebSocket connections
// @Summary WebSocket connection endpoint
// @Description Establishes a WebSocket connection for real-time communication
// @Tags websocket
// @Accept json
// @Produce json
// @Param player_id query string false "Player identifier"
// @Success 101 {string} string "Switching Protocols to websocket"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /ws [get]
func (a *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	a.notifier.HandleWebSocket(w, r)
}
