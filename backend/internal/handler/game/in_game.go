package game

import (
	"log"
	"math/rand"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

func getRandomColor() schema.WoolColor {
	colors := []schema.WoolColor{
		schema.White,     // 0
		schema.Orange,    // 1
		schema.Magenta,   // 2
		schema.LightBlue, // 3
		schema.Yellow,    // 4
		schema.Lime,      // 5
		schema.Pink,      // 6
		schema.Gray,      // 7
		schema.LightGray, // 8
		schema.Cyan,      // 9
		schema.Purple,    // 10
		schema.Blue,      // 11
		schema.Brown,     // 12
		schema.Green,     // 13
		schema.Red,       // 14
		schema.Black,     // 15
	}
	return colors[rand.Intn(len(colors))]
}

// generateRandomMap creates a new random map with all 16 colors
func (h *GameHandler) generateRandomMap(game *schema.Game) {
	for y := 0; y < game.Config.MapHeight; y++ {
		for x := 0; x < game.Config.MapWidth; x++ {
			game.Map[y][x] = getRandomColor()
		}
	}
	log.Printf("Generated new random map for game %s", game.ID)
}

// removeNonTargetColors removes all blocks except the target color, turning them to Air
func (h *GameHandler) removeNonTargetColors(game *schema.Game, targetColor schema.WoolColor) {
	for y := 0; y < game.Config.MapHeight; y++ {
		for x := 0; x < game.Config.MapWidth; x++ {
			if game.Map[y][x] != targetColor {
				game.Map[y][x] = schema.Air
			}
		}
	}
	log.Printf("Removed all non-target colors except %d from game %s", targetColor, game.ID)
}

// calculateRoundDuration returns the rush duration based on round number
func (h *GameHandler) calculateRoundDuration(roundNumber int) float64 {
	// Progressive timing: starts at 4.0s and decreases each round
	// Based on game.md requirement for decreasing countdown each round
	baseDuration := 4.0
	decreasePerRound := 0.1
	minDuration := 1.2

	duration := baseDuration - (float64(roundNumber-1) * decreasePerRound)
	if duration < minDuration {
		duration = minDuration
	}
	return duration
}

func (h *GameHandler) eliminatePlayer(game *schema.Game, player *schema.Player) {
	if player.IsEliminated {
		return
	}

	player.IsEliminated = true
	now := time.Now()
	player.Stats.EliminatedAt = &now
	player.Stats.RoundsSurvived = game.CurrentRound.Number - 1
	// Count alive players for final position
	aliveCount := 0
	for _, p := range game.Players {
		if !p.IsEliminated {
			aliveCount++
		}
	}
	player.Stats.FinalPosition = aliveCount
}

// startNewRound initializes and starts a new round in the game
func (h *GameHandler) startNewRound(game *schema.Game) {
	game.RoundNumber++

	// Step 1: Generate a new map (per game.md requirement)
	h.generateRandomMap(game)

	// Step 2: Determine target color (per game.md requirement)
	targetColor := getRandomColor()

	// Step 3: Calculate progressive round duration (per game.md step 6)
	rushDuration := h.calculateRoundDuration(game.RoundNumber)

	game.CurrentRound = &schema.Round{
		Number:       game.RoundNumber,
		Phase:        schema.ColorCall,
		StartTime:    time.Now(),
		EndTime:      nil,
		ColorToShow:  targetColor,
		RushDuration: rushDuration,
	}

	// Set countdown to rush duration (per game.md step 3)
	game.Countdown = &rushDuration

	log.Printf("Started round %d for game %s with target color %d and duration %.1fs",
		game.RoundNumber, game.ID, targetColor, rushDuration)

	// Broadcast new round start
	game.Broadcast <- map[string]any{
		"event": "round_start",
		"data": map[string]any{
			"round_number": game.RoundNumber,
			"target_color": targetColor,
			"countdown": rushDuration,
			"map": h.convertMapToArray(game),
		},
	}
}

// convertMapToArray converts the map to array format for JSON
func (h *GameHandler) convertMapToArray(game *schema.Game) [][]int {
	mapArray := make([][]int, game.Config.MapHeight)
	for i := range mapArray {
		mapArray[i] = make([]int, game.Config.MapWidth)
		for j := range mapArray[i] {
			mapArray[i][j] = int(game.Map[i][j])
		}
	}
	return mapArray
}

func (h *GameHandler) handleInGamePhase(game *schema.Game) {
	// Ensure there is a current round
	if game.CurrentRound == nil {
		h.startNewRound(game)
		return
	}

	switch game.CurrentRound.Phase {
	case schema.ColorCall:
		h.handleColorCallPhase(game)
	case schema.EliminationCheck:
		h.handleEliminationCheckPhase(game)
	}
}

func (h *GameHandler) handleColorCallPhase(game *schema.Game) {
	// Update countdown timer (per game.md step 3)
	if game.Countdown == nil {
		game.Countdown = &game.CurrentRound.RushDuration
	} else {
		*game.Countdown -= time.Since(game.LastTick).Seconds()
	}

	// Broadcast countdown update
	game.Broadcast <- map[string]any{
		"event": "countdown_update",
		"data": map[string]any{
			"countdown_seconds": game.Countdown,
			"target_color": game.CurrentRound.ColorToShow,
		},
	}

	// When countdown reaches 0, transition to elimination phase
	if game.Countdown == nil || *game.Countdown <= 0 {
		// Step 4: Remove all blocks except target color (per game.md requirement)
		h.removeNonTargetColors(game, game.CurrentRound.ColorToShow)

		// Broadcast map change
		game.Broadcast <- map[string]any{
			"event": "map_update",
			"data": map[string]any{
				"map": h.convertMapToArray(game),
				"blocks_removed": true,
			},
		}

		game.CurrentRound.Phase = schema.EliminationCheck
		game.Countdown = nil
		log.Printf("Round %d countdown finished, removed non-target blocks for game %s",
			game.CurrentRound.Number, game.ID)
	}
}

func (h *GameHandler) handleEliminationCheckPhase(game *schema.Game) {
	eliminatedPlayers := []string{}

	// Step 5: Check each non-eliminated player's position (per game.md requirement)
	for _, player := range game.Players {
		if player.IsEliminated {
			continue
		}

		// Convert player position to map coordinates
		// Player positions are 1-based (1.5, 2.5, etc.), map is 0-based
		x := int(player.Position.X - 1)
		y := int(player.Position.Y - 1)

		// Bounds checking
		if x < 0 || x >= game.Config.MapWidth || y < 0 || y >= game.Config.MapHeight {
			// Player is out of bounds, eliminate them
			h.eliminatePlayer(game, player)
			eliminatedPlayers = append(eliminatedPlayers, player.Name)
			log.Printf("Player %s eliminated (out of bounds) at position (%.1f, %.1f)",
				player.Name, player.Position.X, player.Position.Y)
			continue
		}

		// Check if player is standing on Air (eliminated) or wrong color
		blockUnder := game.Map[y][x]
		if blockUnder == schema.Air || blockUnder != game.CurrentRound.ColorToShow {
			h.eliminatePlayer(game, player)
			eliminatedPlayers = append(eliminatedPlayers, player.Name)
			log.Printf("Player %s eliminated (wrong block: %d, target: %d) at position (%.1f, %.1f)",
				player.Name, blockUnder, game.CurrentRound.ColorToShow, player.Position.X, player.Position.Y)
		}
	}

	// Broadcast elimination results
	if len(eliminatedPlayers) > 0 {
		game.Broadcast <- map[string]any{
			"event": "players_eliminated",
			"data": map[string]any{
				"eliminated_players": eliminatedPlayers,
				"round_number": game.CurrentRound.Number,
				"target_color": game.CurrentRound.ColorToShow,
			},
		}
	}

	// End the current round
	now := time.Now()
	game.CurrentRound.EndTime = &now

	// Count remaining alive players
	aliveCount := 0
	for _, player := range game.Players {
		if !player.IsEliminated {
			aliveCount++
		}
	}
	game.AliveCount = aliveCount

	// Check if game should end (per game.md step 7)
	if aliveCount <= 1 {
		game.Phase = schema.Settlement
		game.EndedAt = &now

		// Find winner if there's exactly one player left
		var winnerID string
		for _, player := range game.Players {
			if !player.IsEliminated {
				winnerID = player.Name
				break
			}
		}

		game.Broadcast <- map[string]any{
			"event": "game_ended",
			"data": map[string]any{
				"winner_id": winnerID,
				"end_time": now,
				"total_rounds": game.RoundNumber,
				"alive_count": aliveCount,
			},
		}

		log.Printf("Game %s ended after %d rounds with winner: %s", game.ID, game.RoundNumber, winnerID)
	} else {
		// Continue to next round (per game.md step 7)
		log.Printf("Round %d completed for game %s, %d players remaining",
			game.CurrentRound.Number, game.ID, aliveCount)

		// Broadcast round end
		game.Broadcast <- map[string]any{
			"event": "round_ended",
			"data": map[string]any{
				"round_number": game.CurrentRound.Number,
				"alive_count": aliveCount,
				"next_round_in": 2.0, // 2 second break between rounds
			},
		}

		// Clear current round and start next one after brief delay
		game.CurrentRound = nil
		game.Countdown = nil

		// Add small delay before next round starts (simulating rest period)
		go func() {
			time.Sleep(2 * time.Second)
			h.startNewRound(game)
		}()
	}
}
