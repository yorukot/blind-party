package game

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yorukot/blind-party/pkg/response"
)

// GetGameState returns the current state of a specific game
func (h *GameHandler) GetGameState(w http.ResponseWriter, r *http.Request) {
	// Extract gameID from URL parameters
	gameID := chi.URLParam(r, "gameID")
	if gameID == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Game ID is required", "MISSING_GAME_ID")
		return
	}

	// Look up the game in GameData map
	game, exists := h.GameData[gameID]
	if !exists {
		response.RespondWithError(w, http.StatusNotFound, "Game not found", "GAME_NOT_FOUND")
		return
	}

	// Return the game state
	response.RespondWithData(w, game)
}