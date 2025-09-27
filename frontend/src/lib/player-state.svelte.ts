import type { PlayerOnBoard, PlayerPosition } from '$lib/types/player';
import type { WebSocketGameClient } from '$lib/api/websocket';
import { SvelteSet } from 'svelte/reactivity';

export type Direction = 'up' | 'down' | 'left' | 'right';

/**
 * Simplified player state management for WebSocket-based movement
 */
class PlayerState {
    activeDirections = $state<SvelteSet<Direction>>(new SvelteSet());
    position = $state<PlayerPosition>({ x: 0, y: 0 });
    playerId = $state<string | null>(null);

    private pressedKeys = new Set<string>();
    private wsClient: WebSocketGameClient | null = null;
    private readonly MOVE_STEP = 1; // Move 1 unit per direction change
    private lastSyncedPosition: PlayerPosition = { x: 0, y: 0 };
    private isInitialized = false;

    private keyDirectionMap: Record<string, Direction> = {
        ArrowUp: 'up',
        ArrowDown: 'down',
        ArrowLeft: 'left',
        ArrowRight: 'right',
        w: 'up',
        s: 'down',
        a: 'left',
        d: 'right'
    };

    private getDirectionFromKey(key: string): Direction | undefined {
        if (!key) {
            return undefined;
        }

        if (this.keyDirectionMap[key]) {
            return this.keyDirectionMap[key];
        }

        const normalized = key.toLowerCase();
        return this.keyDirectionMap[normalized];
    }

    /**
     * Set the WebSocket client for sending position updates
     */
    setWebSocketClient(client: WebSocketGameClient | null): void {
        this.wsClient = client;
    }

    /**
     * Move in a direction and send update to server
     */
    triggerMove(direction: Direction): void {
        console.log(`[PlayerState] triggerMove called with direction: ${direction}`);
        this.activeDirections.add(direction);
        this.moveInDirection(direction);
        this.sendPositionUpdate();
    }

    /**
     * Stop moving in a direction and send update to server
     */
    clearDirection(direction: Direction): void {
        this.activeDirections.delete(direction);
        this.sendPositionUpdate();
    }

    /**
     * Calculate new position based on direction
     */
    private moveInDirection(direction: Direction): void {
        const newPosition = { ...this.position };

        switch (direction) {
            case 'up':
                newPosition.y -= this.MOVE_STEP;
                break;
            case 'down':
                newPosition.y += this.MOVE_STEP;
                break;
            case 'left':
                newPosition.x -= this.MOVE_STEP;
                break;
            case 'right':
                newPosition.x += this.MOVE_STEP;
                break;
        }

        this.position = newPosition;
    }

    /**
     * Send current position to server via WebSocket
     */
    private sendPositionUpdate(): void {
        if (this.wsClient && this.wsClient.getState() === 'connected') {
            console.log(`[PlayerState] Sending position update: (${this.position.x}, ${this.position.y})`);
            this.wsClient.sendPlayerUpdate(this.position.x, this.position.y);
            // Track what we sent to avoid conflicts with server updates
            this.lastSyncedPosition = { x: this.position.x, y: this.position.y };
        } else {
            console.log(`[PlayerState] Cannot send position update - WebSocket state: ${this.wsClient?.getState() || 'null'}`);
        }
    }

    handleKeyDown(event: KeyboardEvent): void {
        const direction = this.getDirectionFromKey(event.key);
        if (!direction) {
            return;
        }

        if (this.pressedKeys.has(event.key)) {
            event.preventDefault();
            return;
        }

        this.pressedKeys.add(event.key);
        event.preventDefault();
        this.triggerMove(direction);
    }

    handleKeyUp(event: KeyboardEvent): void {
        const direction = this.getDirectionFromKey(event.key);
        this.pressedKeys.delete(event.key);
        if (!direction) {
            return;
        }

        this.clearDirection(direction);
    }

    clearPressedKeys(): void {
        this.pressedKeys.clear();
    }

    /**
     * Update position from server game state
     * Only sync if this is the first time or if the server position differs significantly from what we sent
     */
    syncWithServer(player: PlayerOnBoard | null): void {
        if (!player) {
            console.log('[PlayerState] syncWithServer: no player data');
            this.playerId = null;
            this.position = { x: 0, y: 0 };
            this.lastSyncedPosition = { x: 0, y: 0 };
            this.isInitialized = false;
            return;
        }

        // First time initialization
        if (!this.isInitialized || this.playerId !== player.id) {
            console.log(`[PlayerState] syncWithServer: initializing position to (${player.position.x}, ${player.position.y})`);
            this.playerId = player.id;
            this.position = { x: player.position.x, y: player.position.y };
            this.lastSyncedPosition = { x: player.position.x, y: player.position.y };
            this.isInitialized = true;
            return;
        }

        // Check if server position differs significantly from what we last synced
        // This handles cases where the server may have corrected our position
        const distance = Math.abs(player.position.x - this.lastSyncedPosition.x) +
                        Math.abs(player.position.y - this.lastSyncedPosition.y);

        if (distance > 2) { // Only sync if position differs by more than 2 units
            console.log(`[PlayerState] syncWithServer: server correction from (${this.position.x}, ${this.position.y}) to (${player.position.x}, ${player.position.y})`);
            this.position = { x: player.position.x, y: player.position.y };
            this.lastSyncedPosition = { x: player.position.x, y: player.position.y };
        } else {
            console.log(`[PlayerState] syncWithServer: ignoring minor server update (${player.position.x}, ${player.position.y}) - distance: ${distance}`);
        }
    }

    /**
     * Reset player state
     */
    reset(): void {
        this.activeDirections.clear();
        this.position = { x: 0, y: 0 };
        this.playerId = null;
        this.pressedKeys.clear();
        this.wsClient = null;
        this.lastSyncedPosition = { x: 0, y: 0 };
        this.isInitialized = false;
    }
}

export const playerState = new PlayerState();
