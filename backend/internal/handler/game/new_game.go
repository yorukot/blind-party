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
			MapWidth:            256,
			MapHeight:           256,
			CountdownSequence:   []int{30, 25, 20, 15, 10, 8, 6, 4, 3, 2},
			SpectatorOnlyRounds: 2,

			// Timing Progression (rush phase duration by round ranges)
			TimingProgression: []schema.TimingRange{
				{StartRound: 1, EndRound: 3, Duration: 4.0},
				{StartRound: 4, EndRound: 6, Duration: 3.5},
				{StartRound: 7, EndRound: 9, Duration: 3.0},
				{StartRound: 10, EndRound: 12, Duration: 2.5},
				{StartRound: 13, EndRound: 15, Duration: 2.0},
				{StartRound: 16, EndRound: 18, Duration: 1.8},
				{StartRound: 19, EndRound: 21, Duration: 1.5},
				{StartRound: 22, EndRound: 999, Duration: 1.2}, // 22+ rounds
			},

			// Scoring Configuration
			SurvivalPointsPerRound:    10,
			EliminationBonusMultiplier: 5,
			SpeedBonusThreshold:       1.0,
			PerfectBonusThreshold:     2.0,
			SpeedBonusPoints:          2,
			PerfectBonusPoints:        50,
			FinalWinnerBonus:          100,
			EnduranceBonus:            200,
			StreakBonuses:             map[int]int{3: 30, 5: 75, 10: 200},

			// Movement & Anti-cheat
			BaseMovementSpeed:    4.0,
			MaxMovementSpeed:     5.0,
			LagCompensationMs:    100,
			PositionUpdateHz:     10,
			TimerUpdateHz:        20,
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

// generateRandomMap creates a 256x256 map with equal distribution of 16 wool colors
func generateRandomMap() schema.MapData {
	var mapData schema.MapData

	// Create a list of all possible positions
	positions := make([]struct{ x, y int }, 0, 65536) // 256*256 = 65536 total blocks
	for i := 0; i < 256; i++ { // height (rows)
		for j := 0; j < 256; j++ { // width (columns)
			positions = append(positions, struct{ x, y int }{j, i})
		}
	}

	// Shuffle positions for random distribution
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Distribute colors evenly: 16 colors * 4096 blocks = 65536 total blocks (perfect distribution)
	blocksPerColor := 4096 // 65536 / 16 = 4096 blocks per color

	posIndex := 0
	for color := 0; color < 16; color++ {
		// Assign blocks for this color
		for block := 0; block < blocksPerColor; block++ {
			pos := positions[posIndex]
			mapData[pos.y][pos.x] = schema.WoolColor(color)
			posIndex++
		}
	}

	return mapData
}

// mapToArray converts the 2D map array to a format suitable for JSON serialization
func mapToArray(mapData schema.MapData) [][]int {
	result := make([][]int, 256) // height = 256
	for i := 0; i < 256; i++ {
		result[i] = make([]int, 256) // width = 256
		for j := 0; j < 256; j++ {
			result[i][j] = int(mapData[i][j])
		}
	}
	return result
}
