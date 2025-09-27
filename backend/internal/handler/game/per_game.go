package game

import (
	"log"
	"math/rand"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

// handlePreGamePhase manages the pre-game waiting phase
func (h *GameHandler) handlePreGamePhase(game *schema.Game) {
	// Constants from game.md specification
	minPlayers := 4
	maxPlayers := 16

	// Validate player count is within bounds
	if game.PlayerCount > maxPlayers {
		log.Printf("Game %s exceeded maximum players (%d), rejecting new connections", game.ID, maxPlayers)
		return
	}

	// Start game if we have minimum players and no one has joined for a while
	if game.PlayerCount >= minPlayers {
		// Check if we should auto-start (e.g., after waiting period or max players reached)
		if game.PlayerCount >= maxPlayers {
			log.Printf("Game %s reached maximum capacity (%d players), starting immediately", game.ID, maxPlayers)
			h.startGamePreparation(game)
		} else if h.shouldAutoStart(game) {
			log.Printf("Game %s auto-starting with %d players after wait period", game.ID, game.PlayerCount)
			h.startGamePreparation(game)
		}
	}
}

// shouldAutoStart determines if the game should auto-start based on wait time
func (h *GameHandler) shouldAutoStart(game *schema.Game) bool {
	// Auto-start after 30 seconds of having minimum players, or if 75% capacity reached
	timeSinceCreation := time.Since(game.CreatedAt)
	capacityThreshold := float64(game.PlayerCount) / 16.0 // 16 is max players

	return timeSinceCreation > 30*time.Second || capacityThreshold >= 0.75
}

// startGamePreparation begins the 5-second preparation phase
func (h *GameHandler) startGamePreparation(game *schema.Game) {
	log.Printf("Game %s entering preparation phase with %d players", game.ID, game.PlayerCount)

	// Broadcast preparation start
	game.Broadcast <- map[string]interface{}{
		"type": "preparation_started",
		"data": map[string]interface{}{
			"game_id":          game.ID,
			"players":          game.PlayersList,
			"preparation_time": 5, // 5 seconds preparation
		},
	}

	// Start the actual game after 5 seconds
	time.AfterFunc(5*time.Second, func() {
		game.Mu.Lock()
		defer game.Mu.Unlock()
		if game.Phase == schema.PreGame {
			h.startGame(game)
		}
	})
}

// startGame transitions from PreGame to InGame phase
func (h *GameHandler) startGame(game *schema.Game) {
	now := time.Now()
	game.StartedAt = &now
	game.Phase = schema.InGame

	// Assign spawn positions to all players
	h.assignSpawnPositions(game)

	// Initialize player statistics and movement tracking
	h.initializeAllPlayerStats(game)

	log.Printf("Game %s started with %d players", game.ID, game.PlayerCount)

	// Broadcast game start with full game state
	game.Broadcast <- map[string]interface{}{
		"type": "game_started",
		"data": map[string]interface{}{
			"game_id":     game.ID,
			"players":     game.PlayersList,
			"map":         game.MapArray,
			"game_config": game.Config,
		},
	}

	// Start the first round
	h.startNewRound(game)
}

// assignSpawnPositions assigns random spawn positions to all players on valid colored blocks
func (h *GameHandler) assignSpawnPositions(game *schema.Game) {
	// Collect all valid spawn positions (any colored block, not Air)
	validPositions := make([]schema.Position, 0)

	for y := 0; y < game.Config.MapHeight; y++ {
		for x := 0; x < game.Config.MapWidth; x++ {
			if game.Map[y][x] != schema.Air { // Not Air block
				// Use 1-based coordinate system (1-20 range) with 2 decimal precision
				validPositions = append(validPositions, schema.Position{
					X: float64(x+1) + 0.5, // Block coordinates: 1.5, 2.5, ..., 20.5
					Y: float64(y+1) + 0.5,
				})
			}
		}
	}

	// Shuffle positions for random assignment
	rand.Shuffle(len(validPositions), func(i, j int) {
		validPositions[i], validPositions[j] = validPositions[j], validPositions[i]
	})

	// Assign positions to players
	positionIndex := 0
	for _, player := range game.Players {
		if positionIndex < len(validPositions) {
			player.Position = validPositions[positionIndex]
			player.LastValidPosition = player.Position
			positionIndex++

			log.Printf("Player %s (%s) spawned at position (%.1f, %.1f)",
				player.ID, player.Name, player.Position.X, player.Position.Y)
		}
	}
}

// initializeAllPlayerStats initializes statistics and movement tracking for all players
func (h *GameHandler) initializeAllPlayerStats(game *schema.Game) {
	now := time.Now()

	for _, player := range game.Players {
		// Initialize movement tracking
		player.LastUpdate = now
		player.LastMoveTime = now
		player.MovementSpeed = game.Config.BaseMovementSpeed

		// Initialize statistics
		player.Stats = schema.PlayerStats{
			RoundsSurvived:      0,
			TotalDistance:       0,
			Score:               0,
			SurvivalPoints:      0,
			EliminationBonus:    0,
			SpeedBonuses:        0,
			StreakBonuses:       0,
			CurrentStreak:       0,
			LongestStreak:       0,
			ThreeStreakCount:    0,
			FiveStreakCount:     0,
			TenStreakCount:      0,
			AverageResponseTime: 0,
			PerfectRounds:       0,
			FinalPosition:       0,
		}

		log.Printf("Initialized stats for player %s (%s)", player.ID, player.Name)
	}
}
