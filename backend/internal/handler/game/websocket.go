package game

import (
	"fmt"
	"log"
	"strconv"
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
		if msgType, exists := message["event"]; exists {
			switch msgType {
			case "player_update":
				log.Printf("Received player update from user %s", username)
				h.handlePlayerUpdate(game, username, message)
			case "ping":
				// Respond to ping with pong
				client.Send <- map[string]interface{}{
					"event": "pong",
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
		log.Printf("Skipping position update for user %s: player is %s", username,
			func() string {
				if player.IsEliminated { return "eliminated" }
				return "spectator"
			}())
		return
	}

	// Don't allow position updates during elimination phase
	if game.CurrentRound != nil && game.CurrentRound.Phase == schema.EliminationCheck {
		log.Printf("Skipping position update for user %s: game is in elimination phase", username)
		return
	}

	// Extract position data
	data, hasData := message["player"].(map[string]interface{})
	if !hasData {
		log.Printf("Invalid message format from user %s: missing or invalid 'player' field. Message: %+v", username, message)
		return
	}
	log.Printf("Received position data from user %s: %+v", username, data)

	newPosition := player.Position

	// Extract new position coordinates
	if posX, exists := data["pos_x"]; exists {
		if x, err := parseFloat(posX); err == nil {
			newPosition.X = x
			log.Printf("Updated X position for user %s: %.2f", username, x)
		} else {
			log.Printf("Invalid X coordinate from user %s: %v (error: %v)", username, posX, err)
		}
	}

	if posY, exists := data["pos_y"]; exists {
		if y, err := parseFloat(posY); err == nil {
			newPosition.Y = y
			log.Printf("Updated Y position for user %s: %.2f", username, y)
		} else {
			log.Printf("Invalid Y coordinate from user %s: %v (error: %v)", username, posY, err)
		}
	}
	log.Printf("Handling position update for user %s, x: %.1f, y: %.1f", username, newPosition.X, newPosition.Y)

	// Update player position (validation moved to game lifecycle)
	player.Position = newPosition

	// Update last update time
	player.LastUpdate = time.Now()

	game.Players[username] = player
}

// parseFloat attempts to convert various numeric types to float64
func parseFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}
