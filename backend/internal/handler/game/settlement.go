package game

import (
	"log"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// handleSettlementPhase manages the post-game settlement phase
// According to game.md: Settlement phase lasts 5 minutes to show final results
func (h *GameHandler) handleSettlementPhase(game *schema.Game) {
	// Settlement duration as per game.md specification: 5 minutes
	settlementDuration := 5 * time.Minute

	if game.EndedAt == nil {
		log.Printf("Warning: Game %s in settlement phase but EndedAt is nil", game.ID)
		return
	}

	timeInSettlement := time.Since(*game.EndedAt)

	if timeInSettlement >= settlementDuration {
		// Clean up the game after settlement period
		log.Printf("Game %s settlement period completed (5 minutes), cleaning up", game.ID)
		h.cleanupGame(game)
	} else {
		// Continue broadcasting final results periodically during settlement
		h.broadcastFinalResults(game)
	}
}

// broadcastFinalResults sends final game statistics to all connected clients
// This function broadcasts comprehensive final results including leaderboard and game stats
func (h *GameHandler) broadcastFinalResults(game *schema.Game) {
	// Sort players by final position for leaderboard
	sortedPlayers := make([]*schema.Player, len(game.PlayersList))
	copy(sortedPlayers, game.PlayersList)

	// Sort by final position (lower position number = better placement)
	for i := 0; i < len(sortedPlayers); i++ {
		for j := i + 1; j < len(sortedPlayers); j++ {
			if sortedPlayers[i].Stats.FinalPosition > sortedPlayers[j].Stats.FinalPosition {
				sortedPlayers[i], sortedPlayers[j] = sortedPlayers[j], sortedPlayers[i]
			}
		}
	}

	// Calculate game duration
	var gameDuration float64
	if game.StartedAt != nil && game.EndedAt != nil {
		gameDuration = game.EndedAt.Sub(*game.StartedAt).Seconds()
	}

	// Broadcast comprehensive final results
	game.Broadcast <- map[string]interface{}{
		"type": "final_results",
		"data": map[string]interface{}{
			"game_id":      game.ID,
			"total_rounds": len(game.Rounds),
			"duration":     gameDuration,
			"leaderboard":  sortedPlayers,
			"game_stats": map[string]interface{}{
				"total_players":    len(game.Players),
				"rounds_played":    len(game.Rounds),
				"average_survival": h.calculateAverageSurvival(game),
				"winner":           h.determineWinner(game),
				"longest_survival": h.calculateLongestSurvival(game),
			},
		},
	}
}

// calculateAverageSurvival calculates the average rounds survived across all players
func (h *GameHandler) calculateAverageSurvival(game *schema.Game) float64 {
	if len(game.Players) == 0 {
		return 0
	}

	totalSurvival := 0
	for _, player := range game.Players {
		totalSurvival += player.Stats.RoundsSurvived
	}

	return float64(totalSurvival) / float64(len(game.Players))
}

// calculateLongestSurvival finds the maximum rounds survived by any player
func (h *GameHandler) calculateLongestSurvival(game *schema.Game) int {
	maxSurvival := 0
	for _, player := range game.Players {
		if player.Stats.RoundsSurvived > maxSurvival {
			maxSurvival = player.Stats.RoundsSurvived
		}
	}
	return maxSurvival
}

// determineWinner determines the winner based on game.md victory conditions
func (h *GameHandler) determineWinner(game *schema.Game) map[string]interface{} {
	// Primary Victory: Last Player Standing or Multiple Survivors at Round 25
	alivePlayers := make([]*schema.Player, 0)
	for _, player := range game.Players {
		if !player.IsEliminated {
			alivePlayers = append(alivePlayers, player)
		}
	}

	// If only one player alive - solo winner
	if len(alivePlayers) == 1 {
		return map[string]interface{}{
			"type":        "solo_winner",
			"player":      alivePlayers[0],
			"final_score": alivePlayers[0].Stats.Score,
		}
	}

	// If multiple survivors or all eliminated, use secondary victory conditions
	// Tiebreaker #1: Highest Score
	// Tiebreaker #2: Most Rounds Survived
	// Tiebreaker #3: Fastest Average Response Time

	bestPlayers := make([]*schema.Player, 0)
	highestScore := 0

	// Find highest score
	for _, player := range game.Players {
		if player.Stats.Score > highestScore {
			highestScore = player.Stats.Score
		}
	}

	// Get all players with highest score
	for _, player := range game.Players {
		if player.Stats.Score == highestScore {
			bestPlayers = append(bestPlayers, player)
		}
	}

	// If still tied, check rounds survived
	if len(bestPlayers) > 1 {
		mostRounds := 0
		for _, player := range bestPlayers {
			if player.Stats.RoundsSurvived > mostRounds {
				mostRounds = player.Stats.RoundsSurvived
			}
		}

		survivorPlayers := make([]*schema.Player, 0)
		for _, player := range bestPlayers {
			if player.Stats.RoundsSurvived == mostRounds {
				survivorPlayers = append(survivorPlayers, player)
			}
		}
		bestPlayers = survivorPlayers
	}

	// If still tied, check fastest average response time
	if len(bestPlayers) > 1 {
		fastestTime := float64(999999) // Very high initial value
		for _, player := range bestPlayers {
			if player.Stats.AverageResponseTime > 0 && player.Stats.AverageResponseTime < fastestTime {
				fastestTime = player.Stats.AverageResponseTime
			}
		}

		fastestPlayers := make([]*schema.Player, 0)
		for _, player := range bestPlayers {
			if player.Stats.AverageResponseTime == fastestTime {
				fastestPlayers = append(fastestPlayers, player)
			}
		}
		bestPlayers = fastestPlayers
	}

	// Return winner(s)
	if len(bestPlayers) == 1 {
		return map[string]interface{}{
			"type":        "tiebreaker_winner",
			"player":      bestPlayers[0],
			"final_score": bestPlayers[0].Stats.Score,
		}
	} else {
		return map[string]interface{}{
			"type":        "shared_victory",
			"players":     bestPlayers,
			"final_score": highestScore,
		}
	}
}

// cleanupGame removes the game from memory and closes all connections
// This function handles the complete cleanup process when settlement phase ends
func (h *GameHandler) cleanupGame(game *schema.Game) {
	// Stop any running tickers or timers
	if game.Ticker != nil {
		game.Ticker.Stop()
		game.Ticker = nil
	}

	// Send final cleanup notification to all connected clients
	game.Broadcast <- map[string]interface{}{
		"type": "game_cleanup",
		"data": map[string]interface{}{
			"game_id": game.ID,
			"reason":  "settlement_completed",
		},
	}

	// Close all client connections gracefully
	for playerID, client := range game.Clients {
		if client.Conn != nil {
			log.Printf("Closing connection for player %s in game %s", playerID, game.ID)
			client.Conn.Close()
		}
	}

	// Clear all client references
	game.Clients = make(map[string]*schema.WebSocketClient)

	// Remove game from handler's game data
	delete(h.GameData, game.ID)

	log.Printf("Game %s has been cleaned up and removed from memory", game.ID)
}

// transitionToSettlement transitions the game from InGame to Settlement phase
func (h *GameHandler) transitionToSettlement(game *schema.Game) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Set game end time
	now := time.Now()
	game.EndedAt = &now
	game.Phase = schema.Settlement

	// Calculate final positions for any remaining alive players
	h.finalizeFinalPositions(game)

	log.Printf("Game %s transitioned to settlement phase with %d players", game.ID, len(game.Players))

	// Broadcast settlement start
	game.Broadcast <- map[string]interface{}{
		"type": "settlement_started",
		"data": map[string]interface{}{
			"game_id":             game.ID,
			"settlement_duration": 5 * 60, // 5 minutes in seconds
			"final_leaderboard":   h.getFinalLeaderboard(game),
		},
	}

	// Start periodic final results broadcasting during settlement
	h.startSettlementBroadcasting(game)
}

// finalizeFinalPositions ensures all players have proper final positions assigned
func (h *GameHandler) finalizeFinalPositions(game *schema.Game) {
	// Count alive players and assign final positions
	aliveCount := 0
	for _, player := range game.Players {
		if !player.IsEliminated {
			aliveCount++
		}
	}

	// Assign final positions to any remaining alive players
	position := 1
	for _, player := range game.Players {
		if !player.IsEliminated {
			player.Stats.FinalPosition = position
			player.IsEliminated = true // Mark as eliminated for settlement
			if player.Stats.EliminatedAt == nil {
				now := time.Now()
				player.Stats.EliminatedAt = &now
			}
		}
	}

	game.AliveCount = 0
}

// getFinalLeaderboard returns sorted leaderboard for settlement display
func (h *GameHandler) getFinalLeaderboard(game *schema.Game) []*schema.Player {
	sortedPlayers := make([]*schema.Player, len(game.PlayersList))
	copy(sortedPlayers, game.PlayersList)

	// Sort by final position (1 = winner, 2 = second place, etc.)
	for i := 0; i < len(sortedPlayers); i++ {
		for j := i + 1; j < len(sortedPlayers); j++ {
			if sortedPlayers[i].Stats.FinalPosition > sortedPlayers[j].Stats.FinalPosition {
				sortedPlayers[i], sortedPlayers[j] = sortedPlayers[j], sortedPlayers[i]
			}
		}
	}

	return sortedPlayers
}

// startSettlementBroadcasting starts periodic broadcasting of final results during settlement
func (h *GameHandler) startSettlementBroadcasting(game *schema.Game) {
	// Broadcast final results every 10 seconds during settlement
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check if still in settlement phase
				if game.Phase != schema.Settlement {
					return
				}

				// Check if settlement period has ended
				if game.EndedAt != nil && time.Since(*game.EndedAt) >= 5*time.Minute {
					return
				}

				// Broadcast final results
				h.broadcastFinalResults(game)

			case <-time.After(5*time.Minute + 10*time.Second): // Safety timeout
				return
			}
		}
	}()
}
