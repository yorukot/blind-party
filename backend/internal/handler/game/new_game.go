package game

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/yorukot/blind-party/internal/schema"
	"github.com/yorukot/blind-party/pkg/response"
)

func (h *GameHandler) NewGame(w http.ResponseWriter, r *http.Request) {
	// Generate a new 6-digit game ID
	var gameID string
	for {
		// Generate random number between 100000 and 999999
		randomNum := rand.Intn(900000) + 100000
		gameID = strconv.Itoa(randomNum)

		// Check if the game ID already exists
		if _, exists := h.GameData[gameID]; !exists {
			break
		}
	}

	// Create a new game instance
	now := time.Now()
	game := &schema.Game{
		ID:        gameID,
		CreatedAt: now,
		Phase:     schema.PreGame,

		// Initialize maps and slices
		Players:     make(map[string]*schema.Player),
		PlayersList: make([]*schema.Player, 0),
		PlayerCount: 0,
		AliveCount:  0,

		// WebSocket management
		Clients:    make(map[string]*schema.WebSocketClient),
		Broadcast:  make(chan interface{}, 256),
		Register:   make(chan *schema.WebSocketClient, 256),
		Unregister: make(chan *schema.WebSocketClient, 256),

		// Configuration
		Config: schema.GameConfig{
			MapSize:             256,
			CountdownSequence:   []int{30, 25, 20, 15, 10, 8, 6, 4, 3, 2},
			SpectatorOnlyRounds: 2,
		},

		// Initialize rounds slice
		Rounds: make([]schema.Round, 0),

		// Generate random map data
		Map: generateRandomMap(),

		// Synchronization
		StopTicker: make(chan bool),
	}

	// Convert map to array for JSON serialization
	game.MapArray = mapToArray(game.Map)

	// Store the game in GameData map
	h.GameData[gameID] = game

	// Start the game lifecycle in a separate goroutine
	go h.GameLifeCycle(game)

	// Respond with the game ID
	response.RespondWithData(
		w,
		map[string]string{"game_id": gameID},
	)
}

// generateRandomMap creates a 256x256 map with random wool colors
func generateRandomMap() schema.MapData {
	var mapData schema.MapData
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			mapData[i][j] = schema.WoolColor(rand.Intn(16)) // 0-15 wool colors
		}
	}
	return mapData
}

// mapToArray converts the 2D map array to a format suitable for JSON serialization
func mapToArray(mapData schema.MapData) [][]int {
	result := make([][]int, 256)
	for i := 0; i < 256; i++ {
		result[i] = make([]int, 256)
		for j := 0; j < 256; j++ {
			result[i][j] = int(mapData[i][j])
		}
	}
	return result
}
