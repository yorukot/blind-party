package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"

	"github.com/yorukot/blind-party/internal/schema"
)

// Maximum allowed movement speed (blocks per second)
// Adjust this value based on your game's movement mechanics
const MaxMovementSpeed = 0.07

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

	// Generate a unique user ID for this connection
	userID := generateUserID()

	// Create WebSocket client
	client := &schema.WebSocketClient{
		Conn:      ws,
		UserID:    userID,
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
		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					return
				}
				if err := websocket.JSON.Send(ws, message); err != nil {
					log.Printf("Error sending message to client %s: %v", userID, err)
					return
				}
			}
		}
	}()

	// Read messages from client (handle player updates)
	for {
		var message map[string]interface{}
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Printf("WebSocket read error for user %s (username: %s): %v", userID, username, err)
			break
		}

		// Handle different message types
		if msgType, exists := message["type"]; exists {
			switch msgType {
			case "player_update":
				h.handlePlayerUpdate(game, userID, message)
			case "ping":
				// Respond to ping with pong
				client.Send <- map[string]interface{}{
					"type": "pong",
				}
			default:
				log.Printf("Unknown message type from user %s: %s", userID, msgType)
			}
		}
	}
}

// handlePlayerUpdate processes player position updates from WebSocket clients
func (h *GameHandler) handlePlayerUpdate(game *schema.Game, userID string, message map[string]interface{}) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Find the player
	player, exists := game.Players[userID]
	if !exists {
		log.Printf("Player update from unknown user %s", userID)
		return
	}

	// Don't update eliminated or spectator players
	if player.IsEliminated || player.IsSpectator {
		return
	}

	// Don't allow position updates during elimination phase
	if game.CurrentRound != nil && game.CurrentRound.Phase == schema.Eliminating {
		return
	}

	// Extract position data
	data, hasData := message["data"].(map[string]interface{})
	if !hasData {
		return
	}

	// Store old position for speed validation
	oldPosition := player.Position
	oldTime := player.LastUpdate
	newPosition := player.Position

	// Extract new position coordinates
	if posX, exists := data["pos_x"]; exists {
		if x, ok := posX.(float64); ok {
			// Clamp position to map bounds
			if x < 0 {
				x = 0
			} else if x >= 256 {
				x = 255
			}
			newPosition.X = x
		}
	}

	if posY, exists := data["pos_y"]; exists {
		if y, ok := posY.(float64); ok {
			// Clamp position to map bounds
			if y < 0 {
				y = 0
			} else if y >= 256 {
				y = 255
			}
			newPosition.Y = y
		}
	}

	// TODO: we should move this to the game loop? YEAH
	// Validate movement speed using Pythagorean theorem
	currentTime := time.Now()
	timeDelta := currentTime.Sub(oldTime).Seconds()

	// Skip validation for the first update (no previous position)
	if timeDelta > 0 && !oldTime.IsZero() {
		// Calculate distance moved using Pythagorean theorem: sqrt((x2-x1)² + (y2-y1)²)
		deltaX := newPosition.X - oldPosition.X
		deltaY := newPosition.Y - oldPosition.Y
		distance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

		// Calculate actual speed (blocks per second)
		speed := distance / timeDelta

		// Check if speed exceeds maximum allowed
		if speed > MaxMovementSpeed {
			log.Printf("Player %s (%s) moving too fast: %.2f blocks/second (max: %.2f). Rejecting movement.",
				player.ID, player.Name, speed, MaxMovementSpeed)

			// Reject the movement by not updating the position
			// Optionally send a warning to the client
			if client, exists := game.Clients[userID]; exists {
				client.Send <- map[string]interface{}{
					"type": "movement_rejected",
					"data": map[string]interface{}{
						"reason": "movement_too_fast",
						"speed":  speed,
						"max_speed": MaxMovementSpeed,
					},
				}
			}
			return
		}
	}

	// Update player position (movement is valid)
	player.Position = newPosition

	// Update last update time
	player.LastUpdate = currentTime
}

// generateUserID creates a unique user ID
func generateUserID() string {
	return fmt.Sprintf("%d_%d", time.Now().Second(), rand.Intn(10000))
}
