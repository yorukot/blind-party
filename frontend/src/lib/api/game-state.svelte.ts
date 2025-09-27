import { WoolColor } from '$lib/constants/blockTextures';
import type {
    APIClientConfig,
    APIPlayer,
    Game,
    GameConfig,
    Round,
    WebSocketConnectionState
} from '$lib/types/api';
import type { PlayerOnBoard, PlayerSummary } from '$lib/types/player';
import { apiToGrid, gridToApi } from '$lib/utils/coordinates';
import { HTTPClient, createHTTPClient } from './http-client';
import { WebSocketClient, createWebSocketClient } from './websocket-client.svelte';

/**
 * Global game state management for the Color Rush Survival game.
 * Handles API communication, game state synchronization, and reactive updates.
 */
class GameState {
    // Connection and clients
    private httpClient: HTTPClient | null = null;
    private wsClient: WebSocketClient | null = null;

    // Connection state
    connectionState = $state<WebSocketConnectionState>('disconnected');
    lastError = $state<string | null>(null);

    // Game data
    game = $state<Game | null>(null);
    gameId = $state<string | null>(null);
    username = $state<string | null>(null);

    // Current round information
    currentRound = $state<Round | null>(null);
    currentPhase = $state<string | null>(null);
    remainingTime = $state<number>(0);
    currentColor = $state<WoolColor | null>(null);

    // Players (converted to frontend format)
    players = $state<PlayerSummary[]>([]);
    localPlayer = $state<PlayerOnBoard | null>(null);
    localPlayerId = $state<string | null>(null);

    // Map data
    gameMap = $state<number[][]>([]);
    mapSize = $state<number>(20);

    // Game statistics
    playerCount = $state<number>(0);
    aliveCount = $state<number>(0);
    eliminatedPlayers = $state<APIPlayer[]>([]);

    // Configuration
    private config: APIClientConfig = {
        apiBaseUrl: 'http://localhost:8080',
        wsBaseUrl: 'ws://localhost:8080',
        reconnectAttempts: 5,
        reconnectDelay: 1000,
        pingInterval: 30000
    };

    /**
     * Initialize the game state with configuration.
     */
    initialize(customConfig?: Partial<APIClientConfig>): void {
        this.config = { ...this.config, ...customConfig };
        this.httpClient = createHTTPClient(this.config);

        console.log('GameState initialized with config:', this.config);
    }

    /**
     * Create a new game and get the game ID.
     */
    async createGame(): Promise<string> {
        if (!this.httpClient) {
            throw new Error('Game state not initialized');
        }

        try {
            const response = await this.httpClient.createGame();
            this.gameId = response.game_id;
            return response.game_id;
        } catch (error) {
            this.lastError = error instanceof Error ? error.message : 'Failed to create game';
            throw error;
        }
    }

    /**
     * Join a game with the specified game ID and username.
     */
    async joinGame(gameId: string, username: string): Promise<void> {
        try {
            this.gameId = gameId;
            this.username = username;
            this.lastError = null;

            // Create WebSocket client and connect
            this.wsClient = createWebSocketClient(gameId, username, this.config);
            this.setupWebSocketHandlers();

            await this.wsClient.connect();
        } catch (error) {
            this.lastError = error instanceof Error ? error.message : 'Failed to join game';
            throw error;
        }
    }

    /**
     * Disconnect from the current game.
     */
    disconnect(): void {
        if (this.wsClient) {
            this.wsClient.disconnect();
            this.wsClient = null;
        }

        this.resetGameState();
    }

    /**
     * Send a player position update to the server.
     */
    updatePlayerPosition(gridX: number, gridY: number): void {
        if (!this.wsClient?.isConnected) {
            return;
        }

        const apiPosition = gridToApi({ x: gridX, y: gridY });
        this.wsClient.sendPlayerUpdate(apiPosition.pos_x, apiPosition.pos_y);
    }

    /**
     * Get the current game phase.
     */
    get phase(): string {
        return this.game?.phase ?? 'pre-game';
    }

    /**
     * Check if the game is currently active.
     */
    get isGameActive(): boolean {
        return this.game?.phase === 'in-game';
    }

    /**
     * Check if connected to WebSocket.
     */
    get isConnected(): boolean {
        return this.connectionState === 'connected';
    }

    /**
     * Set up WebSocket message handlers.
     */
    private setupWebSocketHandlers(): void {
        if (!this.wsClient) return;

        // Connection state changes
        this.wsClient.onConnectionStateChange((state) => {
            this.connectionState = state;
        });

        // Game state updates
        this.wsClient.on('game_state', (data: Game) => {
            this.updateGameState(data);
        });

        // Player joined
        this.wsClient.on('player_joined', (data: { player: APIPlayer; player_count: number }) => {
            this.playerCount = data.player_count;
            this.addOrUpdatePlayer(data.player);
        });

        // Game started
        this.wsClient.on(
            'game_started',
            (data: { players: APIPlayer[]; map: number[][]; game_config: GameConfig }) => {
                this.updateGameState({
                    ...this.game!,
                    phase: 'in-game',
                    players: data.players,
                    map: data.map,
                    config: data.game_config
                });
            }
        );

        // Color called
        this.wsClient.on(
            'color_called',
            (data: {
                round_number: number;
                color_to_show: WoolColor;
                phase: 'color-call';
                phase_duration: number;
            }) => {
                this.currentColor = data.color_to_show;
                this.currentPhase = 'color-call';
                this.updateCurrentRound(data);
            }
        );

        // Rush phase started
        this.wsClient.on(
            'rush_phase_started',
            (data: { phase: 'rush-phase'; rush_duration: number; round_number: number }) => {
                this.currentPhase = 'rush-phase';
                this.remainingTime = data.rush_duration;
                this.updateCurrentRound(data);
            }
        );

        // Rush timer update
        this.wsClient.on(
            'rush_timer_update',
            (data: { remaining_time: number; round_number: number }) => {
                this.remainingTime = data.remaining_time;
            }
        );

        // Elimination check started
        this.wsClient.on(
            'elimination_check_started',
            (data: { phase: 'elimination-check'; round_number: number }) => {
                this.currentPhase = 'elimination-check';
                this.updateCurrentRound(data);
            }
        );

        // Players eliminated
        this.wsClient.on(
            'players_eliminated',
            (data: {
                eliminated_players: APIPlayer[];
                remaining_count: number;
                round_number: number;
            }) => {
                this.eliminatedPlayers = [...this.eliminatedPlayers, ...data.eliminated_players];
                this.aliveCount = data.remaining_count;

                // Update player statuses
                data.eliminated_players.forEach((eliminatedPlayer: APIPlayer) => {
                    this.updatePlayerStatus(eliminatedPlayer.user_id, 'eliminated');
                });
            }
        );

        // Round results
        this.wsClient.on(
            'round_results',
            (data: {
                phase: 'round-transition';
                round_number: number;
                eliminated_count: number;
                remaining_count: number;
            }) => {
                this.currentPhase = 'round-transition';
                this.aliveCount = data.remaining_count;
            }
        );

        // Player positions update
        this.wsClient.on(
            'player_positions_update',
            (data: {
                players: {
                    user_id: string;
                    name: string;
                    pos_x: number;
                    pos_y: number;
                    is_spectator: boolean;
                }[];
                round_number: number;
                timestamp: number;
            }) => {
                data.players.forEach(
                    (playerUpdate: {
                        user_id: string;
                        name: string;
                        pos_x: number;
                        pos_y: number;
                        is_spectator: boolean;
                    }) => {
                        this.updatePlayerPosition_Internal(
                            playerUpdate.user_id,
                            playerUpdate.pos_x,
                            playerUpdate.pos_y
                        );
                    }
                );
            }
        );

        // Movement rejected
        this.wsClient.on(
            'movement_rejected',
            (data: {
                reason: string;
                speed?: number;
                max_speed?: number;
                reset_position: { pos_x: number; pos_y: number };
                message: string;
            }) => {
                if (this.localPlayer) {
                    const gridPosition = apiToGrid(data.reset_position);
                    this.localPlayer = {
                        ...this.localPlayer,
                        position: gridPosition
                    };
                }
                this.lastError = data.message;
            }
        );

        // Game ended
        this.wsClient.on(
            'game_ended',
            (data: {
                game_id: string;
                total_rounds: number;
                duration: number;
                players: APIPlayer[];
            }) => {
                this.updateGameState({
                    ...this.game!,
                    phase: 'settlement',
                    ended_at: new Date().toISOString(),
                    players: data.players
                });
            }
        );
    }

    /**
     * Update the entire game state.
     */
    private updateGameState(gameData: Game): void {
        this.game = gameData;
        this.gameMap = gameData.map;
        this.playerCount = gameData.player_count;
        this.aliveCount = gameData.alive_count;
        this.currentRound = gameData.current_round ?? null;

        // Convert API players to frontend format
        this.players = gameData.players.map(this.convertAPIPlayerToSummary);

        // Find and set local player
        if (this.username) {
            const localAPIPlayer = gameData.players.find((p) => p.name === this.username);
            if (localAPIPlayer) {
                this.localPlayer = this.convertAPIPlayerToOnBoard(localAPIPlayer);
                this.localPlayerId = localAPIPlayer.user_id;
            }
        }
    }

    /**
     * Add or update a player in the list.
     */
    private addOrUpdatePlayer(apiPlayer: APIPlayer): void {
        const summary = this.convertAPIPlayerToSummary(apiPlayer);
        const existingIndex = this.players.findIndex((p) => p.id === summary.id);

        if (existingIndex >= 0) {
            this.players[existingIndex] = summary;
        } else {
            this.players = [...this.players, summary];
        }

        // Update local player if this is us
        if (this.username && apiPlayer.name === this.username) {
            this.localPlayer = this.convertAPIPlayerToOnBoard(apiPlayer);
            this.localPlayerId = apiPlayer.user_id;
        }
    }

    /**
     * Update player status.
     */
    private updatePlayerStatus(
        playerId: string,
        status: 'spectating' | 'ingame' | 'eliminated'
    ): void {
        const playerIndex = this.players.findIndex((p) => p.id === playerId);
        if (playerIndex >= 0) {
            this.players[playerIndex] = { ...this.players[playerIndex], status };
        }
    }

    /**
     * Update a player's position internally.
     */
    private updatePlayerPosition_Internal(playerId: string, posX: number, posY: number): void {
        const gridPosition = apiToGrid({ pos_x: posX, pos_y: posY });
        const playerIndex = this.players.findIndex((p) => p.id === playerId);

        if (playerIndex >= 0) {
            this.players[playerIndex] = {
                ...this.players[playerIndex],
                position: gridPosition
            };
        }

        // Update local player if this is us
        if (this.localPlayerId === playerId && this.localPlayer) {
            this.localPlayer = {
                ...this.localPlayer,
                position: gridPosition
            };
        }
    }

    /**
     * Update current round information.
     */
    private updateCurrentRound(data: any): void {
        if (this.currentRound) {
            this.currentRound = {
                ...this.currentRound,
                ...data,
                round_number: data.round_number
            };
        }
    }

    /**
     * Convert API player to frontend PlayerSummary.
     */
    private convertAPIPlayerToSummary(apiPlayer: APIPlayer): PlayerSummary {
        return {
            id: apiPlayer.user_id,
            name: apiPlayer.name,
            status: apiPlayer.is_eliminated
                ? 'eliminated'
                : apiPlayer.is_spectator
                  ? 'spectating'
                  : 'ingame',
            accent: this.generatePlayerAccent(apiPlayer.user_id),
            position: apiToGrid(apiPlayer.position)
        };
    }

    /**
     * Convert API player to frontend PlayerOnBoard.
     */
    private convertAPIPlayerToOnBoard(apiPlayer: APIPlayer): PlayerOnBoard {
        const summary = this.convertAPIPlayerToSummary(apiPlayer);
        return {
            ...summary,
            position: summary.position!
        };
    }

    /**
     * Generate a consistent color accent for a player based on their ID.
     */
    private generatePlayerAccent(playerId: string): string {
        const accents = [
            'from-emerald-400 to-emerald-600',
            'from-blue-400 to-indigo-600',
            'from-pink-400 to-rose-600',
            'from-purple-400 to-violet-600',
            'from-yellow-400 to-orange-600',
            'from-green-400 to-teal-600',
            'from-red-400 to-pink-600',
            'from-cyan-400 to-blue-600'
        ];

        // Use a simple hash of the player ID to select an accent
        let hash = 0;
        for (let i = 0; i < playerId.length; i++) {
            hash = ((hash << 5) - hash + playerId.charCodeAt(i)) & 0xffffffff;
        }
        return accents[Math.abs(hash) % accents.length];
    }

    /**
     * Reset all game state.
     */
    private resetGameState(): void {
        this.game = null;
        this.gameId = null;
        this.username = null;
        this.currentRound = null;
        this.currentPhase = null;
        this.remainingTime = 0;
        this.currentColor = null;
        this.players = [];
        this.localPlayer = null;
        this.localPlayerId = null;
        this.gameMap = [];
        this.playerCount = 0;
        this.aliveCount = 0;
        this.eliminatedPlayers = [];
        this.connectionState = 'disconnected';
        this.lastError = null;
    }
}

// Export a singleton instance
export const gameState = new GameState();
