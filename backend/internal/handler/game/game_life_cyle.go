package game

import (
	"log"
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
		log.Printf("Game %s main loop tick", game.ID)
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

	game.Clients[client.Username] = client

	// Determine joined round number
	joinedRound := 0
	if game.CurrentRound != nil {
		joinedRound = game.CurrentRound.Number
	}

	// Create a new player object for this client
	player := &schema.Player{
		Name:              client.Username,
		Position:          schema.Position{X: 10.0, Y: 10.0}, // Default center position
		IsSpectator:       false,
		IsEliminated:      false,
		JoinedRound:       joinedRound,
		LastUpdate:        time.Now(),
		LastValidPosition: schema.Position{X: 10.0, Y: 10.0},
		LastMoveTime:      time.Now(),
		MovementSpeed:     game.Config.BaseMovementSpeed,
		Stats: schema.PlayerStats{
			RoundsSurvived: 0,
			FinalPosition:  0,
		},
	}

	// Add player to the game
	game.Players[client.Username] = player
	game.PlayerCount++
	game.AliveCount++

	log.Printf("Client %s registered to game %s (Player count: %d)", client.Username, game.ID, game.PlayerCount)

	// Send current game state to newly connected client
	gameState := h.createGameStateMessage(game)
	game.Broadcast <- gameState
}

// handleClientUnregister processes WebSocket client disconnections
func (h *GameHandler) handleClientUnregister(game *schema.Game, client *schema.WebSocketClient) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	if _, exists := game.Clients[client.Username]; exists {
		// Remove client
		delete(game.Clients, client.Username)
		close(client.Send)

		// Remove player if it exists
		if player, playerExists := game.Players[client.Username]; playerExists {
			delete(game.Players, client.Username)
			game.PlayerCount--
			// Only decrement alive count if player wasn't eliminated
			if !player.IsEliminated {
				game.AliveCount--
			}
		}

		log.Printf("Client %s unregistered from game %s (Player count: %d)", client.Username, game.ID, game.PlayerCount)

		// Check if no players remain and stop the game
		if game.PlayerCount == 0 {
			log.Printf("No players remaining, stopping game %s", game.ID)
			go func() {
				game.StopTicker <- true
			}()
			return // Don't broadcast since game is stopping
		}

		// Broadcast updated game state to remaining clients via the broadcast channel
		updatedGameState := h.createGameStateMessage(game)
		game.Broadcast <- updatedGameState
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

	// Convert map data to array format for JSON
	game.MapArray = make([][]int, 20)
	for i := range game.MapArray {
		game.MapArray[i] = make([]int, 20)
		for j := range game.MapArray[i] {
			game.MapArray[i][j] = int(game.Map[i][j])
		}
	}

	// Create a safe game state without channels
	return map[string]interface{}{
		"event": "game_update",
		"data": map[string]interface{}{
			"game_id":       game.ID,
			"created_at":    game.CreatedAt,
			"started_at":    game.StartedAt,
			"ended_at":      game.EndedAt,
			"phase":         game.Phase,
			"current_round": game.CurrentRound,
			"map":           game.MapArray,
			"round":         game.CurrentRound,
			"round_number":  game.RoundNumber,
			"players":       game.PlayersList,
			"player_count":  game.PlayerCount,
			"countdown_seconds":     game.Countdown,
			"alive_count":   game.AliveCount,
			"config":        game.Config,
		},
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
		log.Print("Processed InGame phase")
	case schema.Settlement:
		// h.handleSettlementPhase(game)
	}
	game.LastTick = time.Now()
	log.Printf("Game %s state processed (Phase: %s)", game.ID, game.Phase)
	game.Broadcast <- h.createGameStateMessage(game)
}
