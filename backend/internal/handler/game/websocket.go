package game

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"

	"github.com/yorukot/blind-party/internal/schema"
)

// ConnectWebSocket handles WebSocket connections for a specific game
func (h *GameHandler) ConnectWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	// Get gameID from URL path
	req := ws.Request()
	gameID := chi.URLParam(req, "gameID")
	if gameID == "" {
		log.Println("No gameID provided in WebSocket connection")
		return
	}

	// Get game instance
	game, exists := h.GameData[gameID]
	if !exists {
		log.Printf("Game %s not found", gameID)
		return
	}

	// Extract username from query parameters
	username := req.URL.Query().Get("username")
	if username == "" {
		log.Println("No username provided in WebSocket connection")
		return
	}

	// Make sure the username is unique in the game
	for _, player := range game.Players {
		if player.Name == username {
			log.Printf("Username %s already taken in game %s", username, gameID)
			return
		}
	}

	// Create WebSocket client
	client := &schema.WebSocketClient{
		Conn:      ws,
		Username:  username,
		Token:     "", // No token needed
		Send:      make(chan interface{}, 256),
		Connected: time.Now(),
	}

	// Register client with the game
	game.Register <- client

	// Handle client disconnection
	defer func() {
		game.Unregister <- client
	}()

	// Start goroutine to handle sending messages to client
	go func() {
		defer ws.Close()
		for message := range client.Send {
			if err := websocket.JSON.Send(ws, message); err != nil {
				log.Printf("Error sending message to client %s: %v", username, err)
				return
			}
		}
	}()

	// Read messages from client (handle player updates)
	for {
		var message map[string]interface{}
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Printf("WebSocket read error for user %s (username: %s): %v", username, username, err)
			break
		}

		// Handle different message types
		if msgType, exists := message["type"]; exists {
			switch msgType {
			case "player_update":
				h.handlePlayerUpdate(game, username, message)
			case "ping":
				// Respond to ping with pong
				client.Send <- map[string]interface{}{
					"type": "pong",
				}
			default:
				log.Printf("Unknown message type from user %s: %s", username, msgType)
			}
		}
	}
}

// handlePlayerUpdate processes player position updates from WebSocket clients
func (h *GameHandler) handlePlayerUpdate(game *schema.Game, username string, message map[string]interface{}) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Find the player
	player, exists := game.Players[username]
	if !exists {
		log.Printf("Player update from unknown user %s", username)
		return
	}
	// Don't update eliminated or spectator players
	if player.IsEliminated || player.IsSpectator {
		return
	}

	// Don't allow position updates during elimination phase
	if game.CurrentRound != nil && game.CurrentRound.Phase == schema.EliminationCheck {
		return
	}

	// Extract position data
	data, hasData := message["data"].(map[string]interface{})
	if !hasData {
		return
	}

	newPosition := player.Position

	log.Printf("Received position update from user %s: %+v", username, data)
	// Extract new position coordinates
	if posX, exists := data["pos_x"]; exists {
		if x, ok := posX.(float64); ok {
			// Clamp position to map bounds
			if x < 0 {
				x = 0
			} else if x >= 20 {
				x = 20
			}
			newPosition.X = x
			log.Print("Updated X to ", x)
		}
	}

	if posY, exists := data["pos_y"]; exists {
		if y, ok := posY.(float64); ok {
			// Clamp position to map bounds
			if y < 0 {
				y = 0
			} else if y >= 20 {
				y = 20
			}
			newPosition.Y = y
		}
	}

	// Update player position (validation moved to game lifecycle)
	player.Position = newPosition

	// Update last update time
	player.LastUpdate = time.Now()

	// Broadcast position update to all clients
	game.Broadcast <- map[string]interface{}{
		"type": "game_update",
		"data": map[string]interface{}{
			"players": game.Players,
		},
	}
}
