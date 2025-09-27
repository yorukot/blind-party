package game

import (
	"log"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// handlePreGamePhase manages the pre-game waiting phase
func (h *GameHandler) handlePreGamePhase(game *schema.Game) {
	// Start game if we have at least 4 players and 10 seconds have passed
	minPlayers := 4
	waitTime := 10 * time.Second

	if game.PlayerCount >= minPlayers && time.Since(game.CreatedAt) > waitTime {
		h.startGame(game)
	}
}

// startGame transitions from PreGame to InGame phase
func (h *GameHandler) startGame(game *schema.Game) {
	now := time.Now()
	game.StartedAt = &now
	game.Phase = schema.InGame

	// Start the first round
	h.startNewRound(game)

	log.Printf("Game %s started with %d players", game.ID, game.PlayerCount)

	// Broadcast game start
	game.Broadcast <- map[string]interface{}{
		"type": "game_started",
		"data": map[string]interface{}{
			"game_id": game.ID,
			"players": game.PlayersList,
			"round":   game.CurrentRound,
		},
	}
}

