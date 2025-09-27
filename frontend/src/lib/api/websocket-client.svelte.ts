import type {
    ClientMessage,
    ServerMessage,
    WebSocketConnectionState,
    APIClientConfig,
    PlayerUpdateMessage,
    PingMessage
} from '$lib/types/api';

/**
 * WebSocket client for real-time communication with the Color Rush Survival game backend.
 * Handles connection management, message routing, and automatic reconnection.
 */
export class WebSocketClient {
    private ws: WebSocket | null = null;
    private url: string;
    private gameId: string;
    private username: string;
    private reconnectAttempts: number;
    private reconnectDelay: number;
    private pingInterval: number;
    private currentReconnectAttempt = 0;
    private reconnectTimeoutId: number | null = null;
    private pingIntervalId: number | null = null;
    private isManualDisconnect = false;

    private connectionState = $state<WebSocketConnectionState>('disconnected');
    private messageHandlers = new Map<string, Set<(data: any) => void>>();
    private connectionHandlers = new Set<(state: WebSocketConnectionState) => void>();

    constructor(
        gameId: string,
        username: string,
        config: Pick<
            APIClientConfig,
            'wsBaseUrl' | 'reconnectAttempts' | 'reconnectDelay' | 'pingInterval'
        >
    ) {
        this.gameId = gameId;
        this.username = username;
        this.reconnectAttempts = config.reconnectAttempts ?? 5;
        this.reconnectDelay = config.reconnectDelay ?? 1000;
        this.pingInterval = config.pingInterval ?? 30000; // 30 seconds

        const baseUrl = config.wsBaseUrl.endsWith('/')
            ? config.wsBaseUrl.slice(0, -1)
            : config.wsBaseUrl;
        this.url = `${baseUrl}/api/game/${gameId}/ws?username=${encodeURIComponent(username)}`;
    }

    /**
     * Get the current connection state (reactive).
     */
    get state(): WebSocketConnectionState {
        return this.connectionState;
    }

    /**
     * Check if the WebSocket is currently connected.
     */
    get isConnected(): boolean {
        return this.connectionState === 'connected';
    }

    /**
     * Connect to the WebSocket server.
     */
    async connect(): Promise<void> {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            return;
        }

        this.isManualDisconnect = false;
        this.setConnectionState('connecting');

        try {
            await this.establishConnection();
        } catch (error) {
            console.error('Failed to connect to WebSocket:', error);
            this.setConnectionState('error');
            this.scheduleReconnect();
        }
    }

    /**
     * Disconnect from the WebSocket server.
     */
    disconnect(): void {
        this.isManualDisconnect = true;
        this.clearReconnectTimeout();
        this.clearPingInterval();

        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect');
            this.ws = null;
        }

        this.setConnectionState('disconnected');
    }

    /**
     * Send a message to the server.
     */
    send(message: ClientMessage): void {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new WebSocketError('WebSocket is not connected');
        }

        try {
            this.ws.send(JSON.stringify(message));
        } catch (error) {
            console.error('Failed to send message:', error);
            throw new WebSocketError('Failed to send message');
        }
    }

    /**
     * Send a player position update to the server.
     */
    sendPlayerUpdate(posX: number, posY: number): void {
        const message: PlayerUpdateMessage = {
            type: 'player_update',
            data: {
                pos_x: Math.round(posX * 100) / 100, // Round to 2 decimal places
                pos_y: Math.round(posY * 100) / 100
            }
        };
        this.send(message);
    }

    /**
     * Send a ping message to keep the connection alive.
     */
    sendPing(): void {
        const message: PingMessage = {
            type: 'ping'
        };
        this.send(message);
    }

    /**
     * Add a message handler for a specific message type.
     */
    on(messageType: string, handler: (data: any) => void): () => void {
        if (!this.messageHandlers.has(messageType)) {
            this.messageHandlers.set(messageType, new Set());
        }

        const handlers = this.messageHandlers.get(messageType)!;
        handlers.add(handler);

        // Return cleanup function
        return () => {
            handlers.delete(handler);
            if (handlers.size === 0) {
                this.messageHandlers.delete(messageType);
            }
        };
    }

    /**
     * Add a connection state change handler.
     */
    onConnectionStateChange(handler: (state: WebSocketConnectionState) => void): () => void {
        this.connectionHandlers.add(handler);

        // Return cleanup function
        return () => {
            this.connectionHandlers.delete(handler);
        };
    }

    /**
     * Establish a WebSocket connection.
     */
    private async establishConnection(): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.url);

                this.ws.onopen = () => {
                    console.log('WebSocket connected');
                    this.setConnectionState('connected');
                    this.currentReconnectAttempt = 0;
                    this.startPingInterval();
                    resolve();
                };

                this.ws.onmessage = (event) => {
                    try {
                        const message = JSON.parse(event.data) as ServerMessage;
                        this.handleMessage(message);
                    } catch (error) {
                        console.error('Failed to parse WebSocket message:', error);
                    }
                };

                this.ws.onclose = (event) => {
                    console.log('WebSocket closed:', event.code, event.reason);
                    this.clearPingInterval();

                    if (!this.isManualDisconnect) {
                        this.setConnectionState('reconnecting');
                        this.scheduleReconnect();
                    } else {
                        this.setConnectionState('disconnected');
                    }
                };

                this.ws.onerror = (error) => {
                    console.error('WebSocket error:', error);
                    this.setConnectionState('error');
                    reject(new WebSocketError('Connection failed'));
                };
            } catch (error) {
                reject(error);
            }
        });
    }

    /**
     * Handle incoming messages from the server.
     */
    private handleMessage(message: ServerMessage): void {
        const handlers = this.messageHandlers.get(message.type);
        if (handlers) {
            handlers.forEach((handler) => {
                try {
                    handler((message as any).data);
                } catch (error) {
                    console.error(`Error in message handler for ${message.type}:`, error);
                }
            });
        }
    }

    /**
     * Set the connection state and notify handlers.
     */
    private setConnectionState(state: WebSocketConnectionState): void {
        this.connectionState = state;
        this.connectionHandlers.forEach((handler) => {
            try {
                handler(state);
            } catch (error) {
                console.error('Error in connection state handler:', error);
            }
        });
    }

    /**
     * Schedule a reconnection attempt.
     */
    private scheduleReconnect(): void {
        if (this.isManualDisconnect || this.currentReconnectAttempt >= this.reconnectAttempts) {
            this.setConnectionState('error');
            return;
        }

        this.currentReconnectAttempt++;
        const delay = this.reconnectDelay * Math.pow(2, this.currentReconnectAttempt - 1); // Exponential backoff

        console.log(
            `Scheduling reconnect attempt ${this.currentReconnectAttempt}/${this.reconnectAttempts} in ${delay}ms`
        );

        this.reconnectTimeoutId = window.setTimeout(() => {
            this.connect();
        }, delay);
    }

    /**
     * Clear the reconnection timeout.
     */
    private clearReconnectTimeout(): void {
        if (this.reconnectTimeoutId !== null) {
            clearTimeout(this.reconnectTimeoutId);
            this.reconnectTimeoutId = null;
        }
    }

    /**
     * Start the ping interval to keep the connection alive.
     */
    private startPingInterval(): void {
        this.clearPingInterval();
        this.pingIntervalId = window.setInterval(() => {
            if (this.isConnected) {
                try {
                    this.sendPing();
                } catch (error) {
                    console.error('Failed to send ping:', error);
                }
            }
        }, this.pingInterval);
    }

    /**
     * Clear the ping interval.
     */
    private clearPingInterval(): void {
        if (this.pingIntervalId !== null) {
            clearInterval(this.pingIntervalId);
            this.pingIntervalId = null;
        }
    }
}

/**
 * Custom error class for WebSocket errors.
 */
export class WebSocketError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'WebSocketError';
    }
}

/**
 * Default configuration values for the WebSocket client.
 */
export const DEFAULT_WEBSOCKET_CONFIG = {
    wsBaseUrl: 'ws://localhost:8080', // Default backend WebSocket URL
    reconnectAttempts: 5,
    reconnectDelay: 1000, // 1 second
    pingInterval: 30000 // 30 seconds
} as const;

/**
 * Create a WebSocket client instance with default configuration.
 */
export function createWebSocketClient(
    gameId: string,
    username: string,
    customConfig?: Partial<APIClientConfig>
): WebSocketClient {
    const config = {
        wsBaseUrl: customConfig?.wsBaseUrl ?? DEFAULT_WEBSOCKET_CONFIG.wsBaseUrl,
        reconnectAttempts:
            customConfig?.reconnectAttempts ?? DEFAULT_WEBSOCKET_CONFIG.reconnectAttempts,
        reconnectDelay: customConfig?.reconnectDelay ?? DEFAULT_WEBSOCKET_CONFIG.reconnectDelay,
        pingInterval: customConfig?.pingInterval ?? DEFAULT_WEBSOCKET_CONFIG.pingInterval
    };

    return new WebSocketClient(gameId, username, config);
}