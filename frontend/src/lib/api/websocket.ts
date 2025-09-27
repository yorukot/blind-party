import { PUBLIC_WS_BASE_URL } from '$env/static/public';
import type { GameStateResponse, GameEvent } from '../types/game.js';

/**
 * WebSocket connection states
 */
export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'error';

/**
 * WebSocket message types that can be received from the server
 */
export type WebSocketMessage =
	| { event: 'game_update'; data: GameStateResponse }
	| { type: 'error'; data: { message: string; code?: string } }
	| { type: 'ping' }
	| { type: 'pong' };

/**
 * WebSocket message types that can be sent to the server
 */
export type OutgoingWebSocketMessage =
	| { type: 'player_action'; data: { action: string; [key: string]: unknown } }
	| { type: 'ping' }
	| { type: 'pong' };

/**
 * WebSocket client configuration options
 */
export interface WebSocketClientOptions {
	/** Auto-reconnect on connection loss (default: true) */
	autoReconnect?: boolean;
	/** Maximum number of reconnection attempts (default: 5) */
	maxReconnectAttempts?: number;
	/** Delay between reconnection attempts in ms (default: 1000) */
	reconnectDelay?: number;
	/** Ping interval in ms to keep connection alive (default: 30000) */
	pingInterval?: number;
}

/**
 * Event listeners for WebSocket client
 */
export interface WebSocketEventListeners {
	onStateChange?: (state: ConnectionState) => void;
	onGameUpdate?: (gameState: GameStateResponse) => void;
	onError?: (error: string) => void;
	onMessage?: (message: WebSocketMessage) => void;
}

/**
 * WebSocket API client for game connections
 */
export class WebSocketGameClient {
	private ws: WebSocket | null = null;
	private state: ConnectionState = 'disconnected';
	private reconnectAttempts = 0;
	private reconnectTimer: number | null = null;
	private pingTimer: number | null = null;
	private readonly baseUrl: string;
	private gameId: string | null = null;
	private username: string | null = null;

	private readonly options: Required<WebSocketClientOptions>;
	private readonly listeners: WebSocketEventListeners = {};

	constructor(options: WebSocketClientOptions = {}) {
		this.baseUrl = PUBLIC_WS_BASE_URL;
		if (!this.baseUrl) {
			throw new Error('PUBLIC_WS_BASE_URL environment variable is not set');
		}

		this.options = {
			autoReconnect: true,
			maxReconnectAttempts: 5,
			reconnectDelay: 1000,
			pingInterval: 30000,
			...options
		};
	}

	/**
	 * Connect to a specific game
	 */
	async connect(gameId: string, username: string): Promise<void> {
		if (this.state === 'connected' || this.state === 'connecting') {
			throw new Error('WebSocket is already connected or connecting');
		}

		this.gameId = gameId;
		this.username = username;
		this.reconnectAttempts = 0;

		return this.establishConnection();
	}

	/**
	 * Disconnect from the current game
	 */
	disconnect(): void {
		this.clearTimers();
		this.setState('disconnected');

		if (this.ws) {
			this.ws.close(1000, 'Client disconnect');
			this.ws = null;
		}

		this.gameId = null;
		this.username = null;
	}

	/**
	 * Send a message to the server
	 */
	send(message: OutgoingWebSocketMessage): void {
		if (this.state !== 'connected' || !this.ws) {
			throw new Error('WebSocket is not connected');
		}

		this.ws.send(JSON.stringify(message));
	}

	/**
	 * Send a player action
	 */
	sendPlayerAction(action: string, data: Record<string, unknown> = {}): void {
		this.send({
			type: 'player_action',
			data: { action, ...data }
		});
	}

	/**
	 * Get current connection state
	 */
	getState(): ConnectionState {
		return this.state;
	}

	/**
	 * Get current game ID
	 */
	getGameId(): string | null {
		return this.gameId;
	}

	/**
	 * Get current username
	 */
	getUsername(): string | null {
		return this.username;
	}

	/**
	 * Add event listeners
	 */
	on<K extends keyof WebSocketEventListeners>(
		event: K,
		listener: NonNullable<WebSocketEventListeners[K]>
	): void {
		this.listeners[event] = listener;
	}

	/**
	 * Remove event listeners
	 */
	off<K extends keyof WebSocketEventListeners>(event: K): void {
		delete this.listeners[event];
	}

	/**
	 * Establish WebSocket connection
	 */
	private async establishConnection(): Promise<void> {
		if (!this.gameId || !this.username) {
			throw new Error('Game ID and username are required');
		}

		this.setState('connecting');

		const wsUrl = `${this.baseUrl}/api/game/${this.gameId}/ws?username=${encodeURIComponent(this.username)}`;

		try {
			this.ws = new WebSocket(wsUrl);
			this.setupWebSocketEventHandlers();

			// Wait for connection to open or fail
			await new Promise<void>((resolve, reject) => {
				const timeout = setTimeout(() => {
					reject(new Error('Connection timeout'));
				}, 10000);

				this.ws!.onopen = () => {
					clearTimeout(timeout);
					resolve();
				};

				this.ws!.onerror = () => {
					clearTimeout(timeout);
					reject(new Error('WebSocket connection failed'));
				};
			});

		} catch (error) {
			this.setState('error');
			this.listeners.onError?.(
				error instanceof Error ? error.message : 'Unknown connection error'
			);

			if (this.options.autoReconnect && this.reconnectAttempts < this.options.maxReconnectAttempts) {
				this.scheduleReconnect();
			}

			throw error;
		}
	}

	/**
	 * Setup WebSocket event handlers
	 */
	private setupWebSocketEventHandlers(): void {
		if (!this.ws) return;

		this.ws.onopen = () => {
			this.setState('connected');
			this.reconnectAttempts = 0;
			this.startPingTimer();
		};

		this.ws.onclose = (event) => {
			this.clearTimers();

			if (event.code === 1000) {
				// Normal closure
				this.setState('disconnected');
			} else if (this.options.autoReconnect && this.reconnectAttempts < this.options.maxReconnectAttempts) {
				this.setState('reconnecting');
				this.scheduleReconnect();
			} else {
				this.setState('error');
				this.listeners.onError?.(`Connection closed unexpectedly (code: ${event.code})`);
			}
		};

		this.ws.onerror = () => {
			this.setState('error');
			this.listeners.onError?.('WebSocket error occurred');
		};

		this.ws.onmessage = (event) => {
			try {
				const message: WebSocketMessage = JSON.parse(event.data);
				this.handleMessage(message);
			} catch {
				this.listeners.onError?.('Failed to parse WebSocket message');
			}
		};
	}

	/**
	 * Handle incoming WebSocket messages
	 */
	private handleMessage(message: WebSocketMessage): void {
		this.listeners.onMessage?.(message);

		// Handle messages with 'event' property (server format)
		if ('event' in message) {
			switch (message.event) {
				case 'game_update':
					this.listeners.onGameUpdate?.(message.data);
					break;
			}
			return;
		}

		// Handle messages with 'type' property (client/server communication)
		if ('type' in message) {
			switch (message.type) {
				case 'error':
					this.listeners.onError?.(message.data.message);
					break;
				case 'ping':
					this.send({ type: 'pong' });
					break;
				case 'pong':
					// Server acknowledged our ping
					break;
			}
		}
	}

	/**
	 * Set connection state and notify listeners
	 */
	private setState(newState: ConnectionState): void {
		if (this.state !== newState) {
			this.state = newState;
			this.listeners.onStateChange?.(newState);
		}
	}

	/**
	 * Schedule reconnection attempt
	 */
	private scheduleReconnect(): void {
		this.reconnectAttempts++;
		this.reconnectTimer = window.setTimeout(() => {
			if (this.gameId && this.username) {
				this.establishConnection().catch(() => {
					// Error handling is done in establishConnection
				});
			}
		}, this.options.reconnectDelay * this.reconnectAttempts);
	}

	/**
	 * Start ping timer to keep connection alive
	 */
	private startPingTimer(): void {
		this.pingTimer = window.setInterval(() => {
			if (this.state === 'connected') {
				this.send({ type: 'ping' });
			}
		}, this.options.pingInterval);
	}

	/**
	 * Clear all timers
	 */
	private clearTimers(): void {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.pingTimer) {
			clearInterval(this.pingTimer);
			this.pingTimer = null;
		}
	}
}

/**
 * Create a new WebSocket game client instance
 */
export function createWebSocketClient(options?: WebSocketClientOptions): WebSocketGameClient {
	return new WebSocketGameClient(options);
}