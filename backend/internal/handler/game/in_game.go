package game

import (
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


// getRushDurationForRound returns the rush duration based on the round number using timing progression
func (h *GameHandler) getRushDurationForRound(game *schema.Game, roundNumber int) float64 {
	// Use timing progression from config
	for _, timing := range game.Config.TimingProgression {
		if roundNumber >= timing.StartRound && roundNumber <= timing.EndRound {
			return timing.Duration
		}
	}

	// Default to last timing if round exceeds configured ranges
	if len(game.Config.TimingProgression) > 0 {
		lastTiming := game.Config.TimingProgression[len(game.Config.TimingProgression)-1]
		return lastTiming.Duration
	}

	// Fallback if no timing progression configured
	return 2.0
}

// removeNonTargetBlocks converts all blocks that are not the target color to Air
func (h *GameHandler) removeNonTargetBlocks(game *schema.Game) {
	targetColor := game.CurrentRound.ColorToShow
	removedCount := 0

	// Iterate through the entire map
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if game.Map[y][x] != targetColor && game.Map[y][x] != schema.Air {
				game.Map[y][x] = schema.Air
				removedCount++
			}
		}
	}

	// Update the map array for JSON serialization
	game.MapArray = mapToArray(game.Map)
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
	game.Mu.Lock()
	defer game.Mu.Unlock()

	game.RoundNumber++

	// Step 1: Generate a new map
	game.Map = generateRandomMap()
	game.MapArray = mapToArray(game.Map)

	// Step 2: Determine target color
	targetColor := getRandomColor()

	// Step 3: Get rush duration based on round number
	rushDuration := h.getRushDurationForRound(game, game.RoundNumber)

	game.CurrentRound = &schema.Round{
		Number:       game.RoundNumber,
		Phase:        schema.ColorCall,
		StartTime:    time.Now(),
		EndTime:      nil,
		ColorToShow:  targetColor,
		RushDuration: rushDuration,
	}

	// Set countdown to rush duration
	game.Countdown = &rushDuration
}

func (h *GameHandler) handleInGamePhase(game *schema.Game) {
	// Ensure there is a current round
	if game.CurrentRound == nil {
		h.startNewRound(game)
		return
	}

	// Handle the current round phase
	switch game.CurrentRound.Phase {
	case schema.ColorCall:
		h.handleColorCallPhase(game)
	case schema.EliminationCheck:
		h.handleEliminationCheckPhase(game)
		// After elimination check, the round will either end the game or start a new round
		// This is handled within handleEliminationCheckPhase
	}
}

func (h *GameHandler) handleColorCallPhase(game *schema.Game) {
	// Initialize countdown if not set
	if game.Countdown == nil {
		rushDuration := game.CurrentRound.RushDuration
		game.Countdown = &rushDuration
		game.LastTick = time.Now()
	} else {
		// Subtract elapsed time since last tick
		elapsed := time.Since(game.LastTick).Seconds()
		*game.Countdown -= elapsed
		game.LastTick = time.Now()
	}


	// Check if countdown has expired
	if *game.Countdown <= 0 {
		// Transition to elimination check phase
		game.CurrentRound.Phase = schema.EliminationCheck
		game.Countdown = nil
		game.LastTick = time.Now()
	}
}

func (h *GameHandler) handleEliminationCheckPhase(game *schema.Game) {
	// Step 1: Convert all non-target colored blocks to Air (block removal)
	h.removeNonTargetBlocks(game)

	// Step 2: Check each non-eliminated player's position for Air blocks
	for _, player := range game.Players {
		if player.IsEliminated {
			continue
		}

		// Convert player position to map coordinates (1-based to 0-based)
		x := int(player.Position.X) - 1
		y := int(player.Position.Y) - 1

		// Ensure coordinates are within bounds
		if x < 0 || x >= 20 || y < 0 || y >= 20 {
			// Player is out of bounds, eliminate them
			h.eliminatePlayer(game, player)
			continue
		}

		// Check if player is standing on Air (eliminated block)
		if game.Map[y][x] == schema.Air {
			h.eliminatePlayer(game, player)
		}
	}


	// End the current round
	now := time.Now()
	game.CurrentRound.EndTime = &now

	// Update alive count
	game.AliveCount = 0
	for _, player := range game.Players {
		if !player.IsEliminated {
			game.AliveCount++
		}
	}

	// Check if game should end (only one or no players left)
	if game.AliveCount <= 1 {
		game.Phase = schema.Settlement
		game.EndedAt = &now

		// Find winner and set their final position if there's exactly one player left
		for _, player := range game.Players {
			if !player.IsEliminated {
				player.Stats.FinalPosition = 1 // Winner gets position 1
				break
			}
		}

	} else {
		// Continue to next round - clear current round and start new one
		game.CurrentRound = nil
		game.Countdown = nil


		// Start the next round
		h.startNewRound(game)
	}
}
