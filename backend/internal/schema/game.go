package schema

import (
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// WoolColor represents the 16 wool colors in Minecraft
type WoolColor int

const (
	White     WoolColor = iota // 0
	Orange                     // 1
	Magenta                    // 2
	LightBlue                  // 3
	Yellow                     // 4
	Lime                       // 5
	Pink                       // 6
	Gray                       // 7
	LightGray                  // 8
	Cyan                       // 9
	Purple                     // 10
	Blue                       // 11
	Brown                      // 12
	Green                      // 13
	Red                        // 14
	Black                      // 15
	Air                        // 16
)

// GamePhase represents the current phase of the game
type GamePhase string

const (
	PreGame    GamePhase = "pre-game"
	InGame     GamePhase = "in-game"
	Settlement GamePhase = "settlement"
)

// RoundPhase represents the phase within a round
type RoundPhase string

const (
	ColorCall        RoundPhase = "color-call"
	EliminationCheck RoundPhase = "elimination-check"
)

// Position represents x,y coordinates
type Position struct {
	X float64 `json:"pos_x"`
	Y float64 `json:"pos_y"`
}

// Player represents a player in the game
type Player struct {
	Name         string    `json:"name"`
	Position     Position  `json:"position"` // For JSON marshaling
	IsSpectator  bool      `json:"is_spectator"`
	IsEliminated bool      `json:"is_eliminated"`
	JoinedRound  int       `json:"joined_round"`
	LastUpdate   time.Time `json:"-"`

	// Movement validation
	LastValidPosition Position  `json:"-"`
	LastMoveTime      time.Time `json:"-"`
	MovementSpeed     float64   `json:"-"` // blocks per second

	// Stats for settlement
	Stats PlayerStats `json:"-"`
}

// PlayerStats tracks player performance
type PlayerStats struct {
	RoundsSurvived int        `json:"rounds_survived"`
	TotalDistance  float64    `json:"total_distance"`
	EliminatedAt   *time.Time `json:"eliminated_at,omitempty"`
	FinalPosition  int        `json:"final_position"`
}

// Round represents a single round in the game
type Round struct {
	Number          int        `json:"round_number"`
	Phase           RoundPhase `json:"phase"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	ColorToShow     WoolColor  `json:"color_to_show"`
	RushDuration    float64    `json:"rush_duration"` // Variable timing by round
}

// MapData represents the 20x20 game map
type MapData [20][20]WoolColor

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	Conn      *websocket.Conn
	Username  string
	Token     string
	Send      chan interface{}
	Connected time.Time
}

// GameConfig holds configuration for the game
type GameConfig struct {
	MapWidth            int   `json:"map_width"`             // 20
	MapHeight           int   `json:"map_height"`            // 20
	CountdownSequence   []int `json:"countdown_sequence"`    // [30, 25, 20, 15, 10, 8, 6, 4, 3, 2]
	SpectatorOnlyRounds int   `json:"spectator_only_rounds"` // Last 2 rounds

	// Timing Progression (rush phase duration by round ranges)
	TimingProgression []TimingRange `json:"timing_progression"`

	// Scoring Configuration
	SurvivalPointsPerRound     int         `json:"survival_points_per_round"`    // 10
	EliminationBonusMultiplier int         `json:"elimination_bonus_multiplier"` // 5
	SpeedBonusThreshold        float64     `json:"speed_bonus_threshold"`        // 1.0 second
	PerfectBonusThreshold      float64     `json:"perfect_bonus_threshold"`      // 2.0 seconds
	SpeedBonusPoints           int         `json:"speed_bonus_points"`           // 2
	PerfectBonusPoints         int         `json:"perfect_bonus_points"`         // 50
	FinalWinnerBonus           int         `json:"final_winner_bonus"`           // 100
	EnduranceBonus             int         `json:"endurance_bonus"`              // 200
	StreakBonuses              map[int]int `json:"streak_bonuses"`               // {3: 30, 5: 75, 10: 200}

	// Movement & Anti-cheat
	BaseMovementSpeed float64 `json:"base_movement_speed"` // 4.0 blocks/second
	MaxMovementSpeed  float64 `json:"max_movement_speed"`  // 5.0 blocks/second
	LagCompensationMs int     `json:"lag_compensation_ms"` // 100ms
	PositionUpdateHz  int     `json:"position_update_hz"`  // 10 Hz
	TimerUpdateHz     int     `json:"timer_update_hz"`     // 20 Hz

	// Map Changes
	MapChangeRounds    []int `json:"map_change_rounds"`     // Rounds when colors are removed
	ColorsToRemoveEach int   `json:"colors_to_remove_each"` // Number of colors to remove per change
}

// TimingRange defines rush duration for specific round ranges
type TimingRange struct {
	StartRound int     `json:"start_round"`
	EndRound   int     `json:"end_round"`
	Duration   float64 `json:"duration"` // in seconds
}

// Game represents the main game structure
type Game struct {
	// Basic Information
	ID        string     `json:"game_id"`
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	// Game State
	Phase        GamePhase `json:"phase"`
	CurrentRound *Round    `json:"current_round,omitempty"`
	RoundNumber  int        `json:"round_number"`
	Map          MapData   `json:"-"`   // Use MapToArray() for JSON
	MapArray     [][]int   `json:"map"` // Flattened map for JSON
	Countdown    *float64      `json:"countdown_seconds,omitempty"`

	// Players
	Players               map[string]*Player  `json:"-"`
	PlayersList           []*Player           `json:"players"` // For JSON marshaling
	PlayerPositionHistory map[string]Position `json:"-"`       // For movement validation
	PlayerCount           int                 `json:"player_count"`
	AliveCount            int                 `json:"alive_count"`

	// WebSocket Management
	Clients    map[string]*WebSocketClient `json:"-"`
	Broadcast  chan interface{}            `json:"-"`
	Register   chan *WebSocketClient       `json:"-"`
	Unregister chan *WebSocketClient       `json:"-"`

	// Configuration
	Config GameConfig `json:"config"`

	// Synchronization
	Mu                    sync.RWMutex
	Ticker                *time.Ticker
	StopTicker            chan bool
	LastTick              time.Time `json:"-"`
	LastPositionBroadcast time.Time `json:"-"` // Tracks when positions were last broadcast
}
