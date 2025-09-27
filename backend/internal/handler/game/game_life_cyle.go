package game

import (
	"log"
	"math/rand"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
)

func (h *GameHandler) GameLifeCycle(game *schema.Game) {
	defer func() {
		if game.Ticker != nil {
			game.Ticker.Stop()
		}
		log.Printf("Game %s lifecycle ended", game.ID)
	}()

	log.Printf("Starting game lifecycle for game %s", game.ID)

	// Main game loop
	for {
		select {
		case <-game.StopTicker:
			log.Printf("Game %s received stop signal", game.ID)
			return

		case client := <-game.Register:
			h.handleClientRegister(game, client)

		case client := <-game.Unregister:
			h.handleClientUnregister(game, client)

		case message := <-game.Broadcast:
			h.broadcastToClients(game, message)

		default:
			// Handle game state progression
			h.processGameState(game)
			time.Sleep(60 * time.Millisecond)
		}
	}
}

// handleClientRegister processes new WebSocket client connections
func (h *GameHandler) handleClientRegister(game *schema.Game, client *schema.WebSocketClient) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	game.Clients[client.UserID] = client
	log.Printf("Client %s registered to game %s", client.UserID, game.ID)

	// Send current game state to newly connected client
	gameState := h.createGameStateMessage(game)
	client.Send <- gameState
}

// handleClientUnregister processes WebSocket client disconnections
func (h *GameHandler) handleClientUnregister(game *schema.Game, client *schema.WebSocketClient) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	if _, exists := game.Clients[client.UserID]; exists {
		delete(game.Clients, client.UserID)
		close(client.Send)
		log.Printf("Client %s unregistered from game %s", client.UserID, game.ID)
	}
}

// broadcastToClients sends a message to all connected clients
func (h *GameHandler) broadcastToClients(game *schema.Game, message interface{}) {
	game.Mu.RLock()
	defer game.Mu.RUnlock()

	for userID, client := range game.Clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			close(client.Send)
			delete(game.Clients, userID)
			log.Printf("Removed unresponsive client %s from game %s", userID, game.ID)
		}
	}
}

// createGameStateMessage creates a complete game state message for clients
func (h *GameHandler) createGameStateMessage(game *schema.Game) map[string]interface{} {
	// Update players list for JSON serialization
	game.PlayersList = make([]*schema.Player, 0, len(game.Players))
	for _, player := range game.Players {
		game.PlayersList = append(game.PlayersList, player)
	}

	return map[string]interface{}{
		"type": "game_state",
		"data": game,
	}
}
// +=====================================================+
// | 				GAME TICK LOGIC						 |
// +=====================================================+

// processGameState handles the main game logic progression
func (h *GameHandler) processGameState(game *schema.Game) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	switch game.Phase {
	case schema.PreGame:
		h.handlePreGamePhase(game)
	case schema.InGame:
		h.handleInGamePhase(game)
	case schema.Settlement:
		h.handleSettlementPhase(game)
	}
}

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
