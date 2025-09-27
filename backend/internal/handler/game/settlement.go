package game

import (
	"log"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// handleSettlementPhase manages the post-game settlement
func (h *GameHandler) handleSettlementPhase(game *schema.Game) {
	// Settlement phase is mostly passive, just maintain WebSocket connections
	// for players to view final results

	// Auto-cleanup after 5 minutes in settlement
	if game.EndedAt != nil && time.Since(*game.EndedAt) > 5*time.Minute {
		log.Printf("Auto-cleaning up game %s after 5 minutes in settlement", game.ID)
		game.StopTicker <- true
	}
}
