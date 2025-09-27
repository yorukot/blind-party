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

