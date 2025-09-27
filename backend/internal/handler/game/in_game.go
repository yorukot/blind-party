package game

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleInGamePhase manages the active game rounds
func (h *GameHandler) handleInGamePhase(game *schema.Game) {
	if game.CurrentRound == nil {
		return
	}

	// Validate player movements every tick
	h.validatePlayerMovements(game)

	// Check if we need to process round timing
	if game.Ticker != nil {
		select {
		case <-game.Ticker.C:
			h.processRoundTiming(game)
		default:
			// No tick available, continue
		}
	}

	// Check for game end conditions
	if game.AliveCount <= 1 {
		h.endGame(game)
	}
}

// startNewRound creates and starts a new round
func (h *GameHandler) startNewRound(game *schema.Game) {
	roundNumber := len(game.Rounds) + 1
	countdownIndex := min(roundNumber-1, len(game.Config.CountdownSequence)-1)
	countdownTime := game.Config.CountdownSequence[countdownIndex]

	// Select random color for this round (excluding Air)
	colorToShow := schema.WoolColor(rand.Intn(16)) // 0-15 (Air is 16)

	round := schema.Round{
		Number:          roundNumber,
		Phase:           schema.CountingDown,
		CountdownTime:   countdownTime,
		StartTime:       time.Now(),
		ColorToShow:     colorToShow,
		EliminatedCount: 0,
	}

	game.CurrentRound = &round
	game.Rounds = append(game.Rounds, round)

	// Set up round timer
	if game.Ticker != nil {
		game.Ticker.Stop()
	}
	game.Ticker = time.NewTicker(1 * time.Second)

	log.Printf("Round %d started in game %s with color %d and %d second countdown",
		roundNumber, game.ID, colorToShow, countdownTime)

	// Broadcast round start
	game.Broadcast <- map[string]interface{}{
		"type": "round_started",
		"data": map[string]interface{}{
			"round":       round,
			"game_phase": game.Phase,
		},
	}
}

// processRoundTiming handles countdown and phase transitions
func (h *GameHandler) processRoundTiming(game *schema.Game) {
	round := game.CurrentRound
	elapsedTime := int(time.Since(round.StartTime).Seconds())

	switch round.Phase {
	case schema.CountingDown:
		remainingTime := round.CountdownTime - elapsedTime
		if remainingTime <= 0 {
			// Start elimination phase
			round.Phase = schema.Eliminating
			log.Printf("Round %d in game %s entered elimination phase", round.Number, game.ID)

			// Broadcast phase change
			game.Broadcast <- map[string]interface{}{
				"type": "phase_change",
				"data": map[string]interface{}{
					"round_phase": round.Phase,
					"color_to_show": round.ColorToShow,
				},
			}

			// Start elimination after a short delay
			time.AfterFunc(2*time.Second, func() {
				h.eliminatePlayers(game)
			})
		} else {
			// Broadcast countdown update
			game.Broadcast <- map[string]interface{}{
				"type": "countdown_update",
				"data": map[string]interface{}{
					"remaining_time": remainingTime,
					"round_number":   round.Number,
				},
			}
		}

	case schema.Eliminating:
		// Elimination is handled by the timer callback, just wait for next round
		if elapsedTime > round.CountdownTime+5 { // 5 second elimination window
			h.finishRound(game)
		}
	}
}

// eliminatePlayers checks player positions and eliminates those not on the target color
func (h *GameHandler) eliminatePlayers(game *schema.Game) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	round := game.CurrentRound
	if round == nil {
		return
	}

	eliminatedPlayers := make([]*schema.Player, 0)

	for _, player := range game.Players {
		if player.IsEliminated || player.IsSpectator {
			continue
		}

		// Check if player is within map bounds
		x := int(player.Position.X)
		y := int(player.Position.Y)

		if x < 0 || x >= 256 || y < 0 || y >= 256 {
			// Player is out of bounds, eliminate them
			h.eliminatePlayer(game, player, round)
			eliminatedPlayers = append(eliminatedPlayers, player)
			continue
		}

		// Check if player is standing on the correct color
		mapColor := game.Map[y][x] // Note: map is [y][x] for row-column access
		if mapColor != round.ColorToShow {
			// Player is not on the correct color, eliminate them
			h.eliminatePlayer(game, player, round)
			eliminatedPlayers = append(eliminatedPlayers, player)
		}
	}

	round.EliminatedCount = len(eliminatedPlayers)

	if len(eliminatedPlayers) > 0 {
		log.Printf("Eliminated %d players in round %d of game %s",
			len(eliminatedPlayers), round.Number, game.ID)

		// Broadcast eliminations
		game.Broadcast <- map[string]interface{}{
			"type": "players_eliminated",
			"data": map[string]interface{}{
				"eliminated_players": eliminatedPlayers,
				"remaining_count":    game.AliveCount,
				"round_number":       round.Number,
			},
		}
	}
}

// eliminatePlayer marks a player as eliminated and updates stats
func (h *GameHandler) eliminatePlayer(game *schema.Game, player *schema.Player, round *schema.Round) {
	player.IsEliminated = true
	game.AliveCount--

	// Update player stats
	now := time.Now()
	player.Stats.EliminatedAt = &now
	player.Stats.RoundsSurvived = round.Number - 1
	player.Stats.FinalPosition = game.AliveCount + 1 // Position based on remaining players

	log.Printf("Player %s (%s) eliminated in round %d of game %s",
		player.ID, player.Name, round.Number, game.ID)
}

// validatePlayerMovements checks all players for illegal movement speeds every tick
func (h *GameHandler) validatePlayerMovements(game *schema.Game) {
	// Maximum allowed movement speed (blocks per second) - moved from websocket.go
	const MaxMovementSpeed = 0.07

	// Store previous positions in a map (since Player struct doesn't have PreviousPosition field)
	if game.PlayerPositionHistory == nil {
		game.PlayerPositionHistory = make(map[string]schema.Position)
	}

	currentTime := time.Now()

	for _, player := range game.Players {
		if player.IsEliminated || player.IsSpectator {
			continue
		}

		// Get previous position from history
		previousPosition, hasPrevious := game.PlayerPositionHistory[player.ID]

		// Skip validation for the first update (no previous position)
		if !hasPrevious || player.LastUpdate.IsZero() {
			// Store current position as previous for next tick
			game.PlayerPositionHistory[player.ID] = player.Position
			continue
		}

		// Calculate time since last update
		timeDelta := currentTime.Sub(player.LastUpdate).Seconds()

		// Skip if no time has passed
		if timeDelta <= 0 {
			continue
		}

		// Calculate distance moved using Pythagorean theorem: sqrt((x2-x1)² + (y2-y1)²)
		deltaX := player.Position.X - previousPosition.X
		deltaY := player.Position.Y - previousPosition.Y
		distance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

		// Calculate actual speed (blocks per second)
		speed := distance / timeDelta

		// Check if speed exceeds maximum allowed
		if speed > MaxMovementSpeed {
			log.Printf("Player %s (%s) moving too fast: %.2f blocks/second (max: %.2f). Resetting position.",
				player.ID, player.Name, speed, MaxMovementSpeed)

			// Reset player to previous valid position
			player.Position = previousPosition

			// Send warning to client
			if client, exists := game.Clients[player.ID]; exists {
				client.Send <- map[string]interface{}{
					"type": "movement_rejected",
					"data": map[string]interface{}{
						"reason": "movement_too_fast",
						"speed":  speed,
						"max_speed": MaxMovementSpeed,
						"reset_position": player.Position,
					},
				}
			}
		} else {
			// Movement is valid, update position history
			game.PlayerPositionHistory[player.ID] = player.Position
		}
	}
}

// finishRound completes the current round and starts the next one or ends the game
func (h *GameHandler) finishRound(game *schema.Game) {
	round := game.CurrentRound
	if round == nil {
		return
	}

	// Mark round as finished
	now := time.Now()
	round.EndTime = &now

	log.Printf("Round %d finished in game %s, %d players remain",
		round.Number, game.ID, game.AliveCount)

	// Check if game should end
	if game.AliveCount <= 1 {
		h.endGame(game)
		return
	}

	// Start next round after a brief pause
	time.AfterFunc(3*time.Second, func() {
		game.Mu.Lock()
		defer game.Mu.Unlock()
		if game.Phase == schema.InGame {
			h.startNewRound(game)
		}
	})

	// Broadcast round end
	game.Broadcast <- map[string]interface{}{
		"type": "round_finished",
		"data": map[string]interface{}{
			"round_number":    round.Number,
			"eliminated_count": round.EliminatedCount,
			"remaining_count":  game.AliveCount,
			"next_round_in":    3,
		},
	}
}

// endGame transitions the game to settlement phase
func (h *GameHandler) endGame(game *schema.Game) {
	game.Phase = schema.Settlement
	now := time.Now()
	game.EndedAt = &now

	// Stop the ticker
	if game.Ticker != nil {
		game.Ticker.Stop()
		game.Ticker = nil
	}

	// Calculate final rankings
	h.calculateFinalRankings(game)

	log.Printf("Game %s ended after %d rounds", game.ID, len(game.Rounds))

	// Broadcast game end
	game.Broadcast <- map[string]interface{}{
		"type": "game_ended",
		"data": map[string]interface{}{
			"game_id":      game.ID,
			"total_rounds": len(game.Rounds),
			"duration":     game.EndedAt.Sub(*game.StartedAt).Seconds(),
			"players":      game.PlayersList,
		},
	}
}

// calculateFinalRankings determines final player positions
func (h *GameHandler) calculateFinalRankings(game *schema.Game) {
	// Find the winner (last player standing)
	for _, player := range game.Players {
		if !player.IsEliminated {
			player.Stats.FinalPosition = 1
			player.Stats.RoundsSurvived = len(game.Rounds)
			log.Printf("Player %s (%s) won game %s", player.ID, player.Name, game.ID)
			break
		}
	}

	// Update players list for final broadcast
	game.PlayersList = make([]*schema.Player, 0, len(game.Players))
	for _, player := range game.Players {
		game.PlayersList = append(game.PlayersList, player)
	}
}
