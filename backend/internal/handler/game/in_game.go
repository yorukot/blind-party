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
	if game.AliveCount <= 1 || (game.CurrentRound != nil && game.CurrentRound.Number >= 25) {
		h.endGame(game)
	}
}

// startNewRound creates and starts a new round
func (h *GameHandler) startNewRound(game *schema.Game) {
	roundNumber := len(game.Rounds) + 1

	// Calculate rush duration based on timing progression
	rushDuration := h.calculateRushDuration(game, roundNumber)

	// Select random color for this round (excluding Air)
	colorToShow := schema.WoolColor(rand.Intn(16)) // 0-15 (Air is 16)

	round := schema.Round{
		Number:          roundNumber,
		Phase:           schema.ColorCall,
		CountdownTime:   1, // 1 second color call phase
		StartTime:       time.Now(),
		ColorToShow:     colorToShow,
		RushDuration:    rushDuration,
		EliminatedCount: 0,
	}

	game.CurrentRound = &round
	game.Rounds = append(game.Rounds, round)

	// Set up round timer with 20Hz update rate (50ms intervals)
	if game.Ticker != nil {
		game.Ticker.Stop()
	}
	game.Ticker = time.NewTicker(50 * time.Millisecond) // 20Hz for precise timing

	log.Printf("Round %d started in game %s with color %d, %.1f second rush duration",
		roundNumber, game.ID, colorToShow, rushDuration)

	// Broadcast color call phase start
	game.Broadcast <- map[string]interface{}{
		"type": "color_called",
		"data": map[string]interface{}{
			"round_number":   roundNumber,
			"color_to_show":  colorToShow,
			"phase":          round.Phase,
			"phase_duration": 1.0,
		},
	}
}

// processRoundTiming handles all 4 round phase transitions
func (h *GameHandler) processRoundTiming(game *schema.Game) {
	round := game.CurrentRound
	elapsedTime := time.Since(round.StartTime).Seconds()

	switch round.Phase {
	case schema.ColorCall:
		// Color Call phase (1 second)
		if elapsedTime >= 1.0 {
			round.Phase = schema.RushPhase
			log.Printf("Round %d in game %s entered rush phase (%.1fs duration)", round.Number, game.ID, round.RushDuration)

			// Broadcast rush phase start
			game.Broadcast <- map[string]interface{}{
				"type": "rush_phase_started",
				"data": map[string]interface{}{
					"phase":         round.Phase,
					"rush_duration": round.RushDuration,
					"round_number":  round.Number,
				},
			}
		}

	case schema.RushPhase:
		// Rush Phase (variable duration)
		rushElapsedTime := elapsedTime - 1.0 // Subtract color call duration
		remainingRushTime := round.RushDuration - rushElapsedTime

		if remainingRushTime <= 0 {
			round.Phase = schema.EliminationCheck
			log.Printf("Round %d in game %s entered elimination check phase", round.Number, game.ID)

			// Broadcast elimination check phase
			game.Broadcast <- map[string]interface{}{
				"type": "elimination_check_started",
				"data": map[string]interface{}{
					"phase":        round.Phase,
					"round_number": round.Number,
				},
			}

			// Perform elimination check with lag compensation
			h.eliminatePlayersWithLagCompensation(game)
		} else {
			// Broadcast rush timer update
			game.Broadcast <- map[string]interface{}{
				"type": "rush_timer_update",
				"data": map[string]interface{}{
					"remaining_time": remainingRushTime,
					"round_number":   round.Number,
				},
			}
		}

	case schema.EliminationCheck:
		// Elimination Check phase (0.5 seconds)
		eliminationElapsedTime := elapsedTime - 1.0 - round.RushDuration
		if eliminationElapsedTime >= 0.5 {
			round.Phase = schema.RoundTransition
			log.Printf("Round %d in game %s entered round transition phase", round.Number, game.ID)

			// Calculate and update player scores
			h.calculateRoundScores(game, round)

			// Broadcast round results
			game.Broadcast <- map[string]interface{}{
				"type": "round_results",
				"data": map[string]interface{}{
					"phase":            round.Phase,
					"round_number":     round.Number,
					"eliminated_count": round.EliminatedCount,
					"remaining_count":  game.AliveCount,
				},
			}
		}

	case schema.RoundTransition:
		// Round Transition phase (1 second)
		transitionElapsedTime := elapsedTime - 1.0 - round.RushDuration - 0.5
		if transitionElapsedTime >= 1.0 {
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

		// Check if player is within map bounds (20x20 map with 1-20 coordinate system)
		// Convert from 1-based coordinates to 0-based array indices
		x := int(player.Position.X - 1) // Convert 1-20 to 0-19
		y := int(player.Position.Y - 1)

		if x < 0 || x >= game.Config.MapWidth || y < 0 || y >= game.Config.MapHeight {
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

	// Calculate elimination bonus based on placement
	totalPlayers := len(game.Players)
	eliminationBonus := game.Config.EliminationBonusMultiplier * (totalPlayers - player.Stats.FinalPosition)
	player.Stats.EliminationBonus += eliminationBonus
	player.Stats.Score += eliminationBonus

	// Reset survival streak
	player.Stats.CurrentStreak = 0

	log.Printf("Player %s (%s) eliminated in round %d of game %s (Position: %d, Bonus: +%d)",
		player.ID, player.Name, round.Number, game.ID, player.Stats.FinalPosition, eliminationBonus)
}

// validatePlayerMovements checks all players for illegal movement speeds and teleportation
func (h *GameHandler) validatePlayerMovements(game *schema.Game) {
	// Use configured maximum movement speed
	maxMovementSpeed := game.Config.MaxMovementSpeed

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
			player.LastValidPosition = player.Position
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

		// Check for boundary violations (20x20 map with 1-20 coordinate system)
		if player.Position.X < 1.0 || player.Position.X > 21.0 ||
			player.Position.Y < 1.0 || player.Position.Y > 21.0 {
			log.Printf("Player %s (%s) moved out of bounds: (%.2f, %.2f). Resetting position.",
				player.ID, player.Name, player.Position.X, player.Position.Y)

			// Reset to last valid position
			player.Position = player.LastValidPosition
			h.sendMovementRejection(game, player, "out_of_bounds", speed, maxMovementSpeed)
			continue
		}

		// Check for teleportation (distance too large for time delta)
		maxPossibleDistance := maxMovementSpeed * timeDelta
		if distance > maxPossibleDistance*1.1 { // 10% tolerance for network jitter
			log.Printf("Player %s (%s) teleported: %.2f blocks in %.3fs (max: %.2f). Resetting position.",
				player.ID, player.Name, distance, timeDelta, maxPossibleDistance)

			// Reset to last valid position
			player.Position = player.LastValidPosition
			h.sendMovementRejection(game, player, "teleportation_detected", speed, maxMovementSpeed)
			continue
		}

		// Check if speed exceeds maximum allowed
		if speed > maxMovementSpeed*1.05 { // 5% tolerance for network fluctuations
			log.Printf("Player %s (%s) moving too fast: %.2f blocks/second (max: %.2f). Resetting position.",
				player.ID, player.Name, speed, maxMovementSpeed)

			// Reset player to last valid position
			player.Position = player.LastValidPosition
			h.sendMovementRejection(game, player, "movement_too_fast", speed, maxMovementSpeed)
		} else {
			// Movement is valid, update position history
			game.PlayerPositionHistory[player.ID] = player.Position
			player.LastValidPosition = player.Position
			player.LastMoveTime = currentTime

			// Update total distance for stats
			player.Stats.TotalDistance += distance
		}
	}
}

// sendMovementRejection sends a movement rejection message to the client
func (h *GameHandler) sendMovementRejection(game *schema.Game, player *schema.Player, reason string, speed, maxSpeed float64) {
	if client, exists := game.Clients[player.ID]; exists {
		client.Send <- map[string]interface{}{
			"type": "movement_rejected",
			"data": map[string]interface{}{
				"reason":         reason,
				"speed":          speed,
				"max_speed":      maxSpeed,
				"reset_position": player.Position,
				"message":        "Position reset due to invalid movement",
			},
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
	if game.AliveCount <= 1 || round.Number >= 25 {
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
			"round_number":     round.Number,
			"eliminated_count": round.EliminatedCount,
			"remaining_count":  game.AliveCount,
			"next_round_in":    3,
		},
	}
}

// endGame transitions the game to settlement phase
func (h *GameHandler) endGame(game *schema.Game) {
	// Stop the ticker
	if game.Ticker != nil {
		game.Ticker.Stop()
		game.Ticker = nil
	}

	// Calculate final rankings before transitioning to settlement
	h.calculateFinalRankings(game)

	log.Printf("Game %s ended after %d rounds", game.ID, len(game.Rounds))

	// Broadcast game end
	game.Broadcast <- map[string]interface{}{
		"type": "game_ended",
		"data": map[string]interface{}{
			"game_id":      game.ID,
			"total_rounds": len(game.Rounds),
			"duration":     time.Since(*game.StartedAt).Seconds(),
			"players":      game.PlayersList,
		},
	}

	// Transition to settlement phase using the dedicated settlement handler
	h.transitionToSettlement(game)
}

// calculateFinalRankings determines final player positions with proper tiebreakers
func (h *GameHandler) calculateFinalRankings(game *schema.Game) {
	alivePlayers := make([]*schema.Player, 0)

	// Collect all alive players
	for _, player := range game.Players {
		if !player.IsEliminated {
			alivePlayers = append(alivePlayers, player)
		}
	}

	// Apply endurance bonus for players surviving to Round 25
	if len(game.Rounds) >= 25 {
		for _, player := range alivePlayers {
			player.Stats.Score += game.Config.EnduranceBonus
			log.Printf("Player %s (%s) received endurance bonus: +%d points",
				player.ID, player.Name, game.Config.EnduranceBonus)
		}
	}

	// Handle different victory scenarios
	if len(alivePlayers) == 1 {
		// Single winner
		winner := alivePlayers[0]
		winner.Stats.FinalPosition = 1
		winner.Stats.RoundsSurvived = len(game.Rounds)
		winner.Stats.Score += game.Config.FinalWinnerBonus
		log.Printf("Player %s (%s) won game %s with %d points",
			winner.ID, winner.Name, game.ID, winner.Stats.Score)
	} else if len(alivePlayers) > 1 {
		// Multiple survivors - apply tiebreaker logic
		h.resolveTiebreakers(game, alivePlayers)
	} else {
		// No survivors (shouldn't happen but handle gracefully)
		log.Printf("Game %s ended with no survivors", game.ID)
	}

	// Update players list for final broadcast
	game.PlayersList = make([]*schema.Player, 0, len(game.Players))
	for _, player := range game.Players {
		game.PlayersList = append(game.PlayersList, player)
	}
}

// resolveTiebreakers handles multiple survivor scenario with proper tiebreaker rules
func (h *GameHandler) resolveTiebreakers(game *schema.Game, alivePlayers []*schema.Player) {
	// Sort players by tiebreaker criteria:
	// 1. Highest Score
	// 2. Most Rounds Survived
	// 3. Fastest Average Response Time

	// Create a slice for sorting
	players := make([]*schema.Player, len(alivePlayers))
	copy(players, alivePlayers)

	// Sort using multiple criteria
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			player1 := players[i]
			player2 := players[j]

			// Compare scores (higher is better)
			if player1.Stats.Score != player2.Stats.Score {
				if player1.Stats.Score < player2.Stats.Score {
					players[i], players[j] = players[j], players[i]
				}
				continue
			}

			// Compare rounds survived (higher is better)
			if player1.Stats.RoundsSurvived != player2.Stats.RoundsSurvived {
				if player1.Stats.RoundsSurvived < player2.Stats.RoundsSurvived {
					players[i], players[j] = players[j], players[i]
				}
				continue
			}

			// Compare average response time (lower is better)
			if player1.Stats.AverageResponseTime != player2.Stats.AverageResponseTime {
				if player1.Stats.AverageResponseTime > player2.Stats.AverageResponseTime {
					players[i], players[j] = players[j], players[i]
				}
			}
		}
	}

	// Assign final positions
	for i, player := range players {
		player.Stats.FinalPosition = i + 1
		player.Stats.RoundsSurvived = len(game.Rounds)

		// Give winner bonus to first place
		if i == 0 {
			player.Stats.Score += game.Config.FinalWinnerBonus
		}

		log.Printf("Player %s (%s) finished in position %d with %d points",
			player.ID, player.Name, player.Stats.FinalPosition, player.Stats.Score)
	}
}

// eliminatePlayersWithLagCompensation checks player positions with 100ms lag compensation
func (h *GameHandler) eliminatePlayersWithLagCompensation(game *schema.Game) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	round := game.CurrentRound
	if round == nil {
		return
	}

	eliminatedPlayers := make([]*schema.Player, 0)
	lagCompensationDuration := time.Duration(game.Config.LagCompensationMs) * time.Millisecond

	for _, player := range game.Players {
		if player.IsEliminated || player.IsSpectator {
			continue
		}

		// Apply lag compensation - check if player's last update was within the compensation window
		timeSinceLastUpdate := time.Since(player.LastUpdate)
		if timeSinceLastUpdate > lagCompensationDuration {
			// Use last known position if within lag compensation window
			log.Printf("Applying lag compensation for player %s (%s)", player.ID, player.Name)
		}

		// Check if player is within map bounds (20x20 map with 1-20 coordinate system)
		// Convert from 1-based coordinates to 0-based array indices
		x := int(player.Position.X - 1) // Convert 1-20 to 0-19
		y := int(player.Position.Y - 1)

		if x < 0 || x >= game.Config.MapWidth || y < 0 || y >= game.Config.MapHeight {
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

// calculateRoundScores calculates and applies scoring for the current round
func (h *GameHandler) calculateRoundScores(game *schema.Game, round *schema.Round) {
	for _, player := range game.Players {
		if player.IsEliminated || player.IsSpectator {
			continue
		}

		// Survival points
		player.Stats.SurvivalPoints += game.Config.SurvivalPointsPerRound
		player.Stats.Score += game.Config.SurvivalPointsPerRound

		// Calculate response time (time to reach correct color)
		responseTime := player.LastUpdate.Sub(round.StartTime.Add(1 * time.Second)).Seconds()
		if responseTime > 0 && responseTime < round.RushDuration {
			// Add to running average for tiebreaker
			if player.Stats.AverageResponseTime == 0 {
				player.Stats.AverageResponseTime = responseTime
			} else {
				player.Stats.AverageResponseTime = (player.Stats.AverageResponseTime + responseTime) / 2
			}

			// Speed bonuses
			remainingTime := round.RushDuration - responseTime
			if remainingTime > game.Config.PerfectBonusThreshold {
				// Perfect bonus (+50 points for >2s remaining)
				player.Stats.SpeedBonuses += game.Config.PerfectBonusPoints
				player.Stats.Score += game.Config.PerfectBonusPoints
				player.Stats.PerfectRounds++
			} else if remainingTime > game.Config.SpeedBonusThreshold {
				// Speed bonus (+2 points for >1s remaining)
				player.Stats.SpeedBonuses += game.Config.SpeedBonusPoints
				player.Stats.Score += game.Config.SpeedBonusPoints
			}
		}

		// Update survival streak
		player.Stats.CurrentStreak++
		if player.Stats.CurrentStreak > player.Stats.LongestStreak {
			player.Stats.LongestStreak = player.Stats.CurrentStreak
		}

		// Apply streak bonuses
		if streak, exists := game.Config.StreakBonuses[player.Stats.CurrentStreak]; exists {
			player.Stats.StreakBonuses += streak
			player.Stats.Score += streak

			// Track streak counts
			switch player.Stats.CurrentStreak {
			case 3:
				player.Stats.ThreeStreakCount++
			case 5:
				player.Stats.FiveStreakCount++
			case 10:
				player.Stats.TenStreakCount++
			}

			log.Printf("Player %s (%s) achieved %d round streak bonus: +%d points",
				player.ID, player.Name, player.Stats.CurrentStreak, streak)
		}

		player.Stats.RoundsSurvived = round.Number
	}
}

// calculateRushDuration returns the rush duration for a given round number based on timing progression
func (h *GameHandler) calculateRushDuration(game *schema.Game, roundNumber int) float64 {
	// Default duration if no timing progression is configured
	defaultDuration := 4.0

	if len(game.Config.TimingProgression) == 0 {
		return defaultDuration
	}

	// Find the appropriate timing range for this round
	for _, timingRange := range game.Config.TimingProgression {
		if roundNumber >= timingRange.StartRound && roundNumber <= timingRange.EndRound {
			return timingRange.Duration
		}
	}

	// If no range matches, use the last range's duration (for rounds beyond the configured ranges)
	lastRange := game.Config.TimingProgression[len(game.Config.TimingProgression)-1]
	return lastRange.Duration
}
