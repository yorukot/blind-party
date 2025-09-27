package game

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// JoinGame handles requests for a player to join the game
func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameID := r.URL.Query().Get("game_id")
	userID := r.URL.Query().Get("user_id")
	name := r.URL.Query().Get("name")

	if gameID == "" || userID == "" || name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing game_id, user_id or name",
		})
		return
	}

	// Get the game instance
	game, exists := h.GameData[gameID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "game not found",
		})
		return
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Check if player already exists
	if _, exists := game.Players[userID]; exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "player already joined",
		})
		return
	}

	// Check if game is still accepting players
	if game.Phase != schema.PreGame {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "game has already started",
		})
		return
	}

	// Create new player
	player := &schema.Player{
		ID:           userID,
		Name:         name,
		Position:     schema.Position{X: 128, Y: 128}, // Start at center of map
		IsSpectator:  false,
		IsEliminated: false,
		JoinedRound:  len(game.Rounds) + 1,
		LastUpdate:   time.Now(),
		Stats:        schema.PlayerStats{},
	}

	// Add player to game
	game.Players[userID] = player
	game.PlayersList = append(game.PlayersList, player)
	game.PlayerCount++
	game.AliveCount++

	// Broadcast player joined to all clients
	game.Broadcast <- map[string]interface{}{
		"type": "player_joined",
		"data": map[string]interface{}{
			"player":       player,
			"player_count": game.PlayerCount,
		},
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "player joined successfully",
		"player":  player,
		"game_id": game.ID,
	})
}
