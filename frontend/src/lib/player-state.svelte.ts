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
    private lastSyncedPosition: PlayerPosition = { x: 0, y: 0 };
    private isInitialized = false;

    // Smooth movement physics
    private velocity = { x: 0, y: 0 };
    private readonly ACCELERATION = 0.8; // Units per frame squared
    private readonly MAX_SPEED = 0.15; // Maximum velocity per frame
    private readonly FRICTION = 0.85; // Friction coefficient (0-1)
    private animationFrameId: number | null = null;
    private lastUpdateTime = 0;
    private lastPositionSent = { x: 0, y: 0 };
    private readonly POSITION_SEND_INTERVAL = 100; // Send position every 100ms
    private lastPositionSentTime = 0;

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
     * Start moving in a direction (smooth acceleration)
     */
    triggerMove(direction: Direction): void {
        this.activeDirections.add(direction);
        this.startMovementLoop();
    }

    /**
     * Stop moving in a direction
     */
    clearDirection(direction: Direction): void {
        this.activeDirections.delete(direction);
        // Movement loop will handle deceleration when no directions are active
    }

    /**
     * Start the smooth movement animation loop
     */
    private startMovementLoop(): void {
        if (this.animationFrameId !== null) {
            return; // Already running
        }

        this.lastUpdateTime = performance.now();
        this.updateMovement();
    }

    /**
     * Update player position with smooth physics
     */
    private updateMovement = (): void => {
        const currentTime = performance.now();
        const deltaTime = Math.min(currentTime - this.lastUpdateTime, 32) / 16.67; // Cap at ~60fps, normalize to 60fps
        this.lastUpdateTime = currentTime;

        // Calculate target velocity based on active directions
        const targetVelocity = { x: 0, y: 0 };

        if (this.activeDirections.has('left')) targetVelocity.x -= 1;
        if (this.activeDirections.has('right')) targetVelocity.x += 1;
        if (this.activeDirections.has('up')) targetVelocity.y -= 1;
        if (this.activeDirections.has('down')) targetVelocity.y += 1;

        // Normalize diagonal movement
        const magnitude = Math.sqrt(targetVelocity.x * targetVelocity.x + targetVelocity.y * targetVelocity.y);
        if (magnitude > 0) {
            targetVelocity.x = (targetVelocity.x / magnitude) * this.MAX_SPEED;
            targetVelocity.y = (targetVelocity.y / magnitude) * this.MAX_SPEED;
        }

        // Apply acceleration towards target velocity
        const accel = this.ACCELERATION * deltaTime;
        this.velocity.x += (targetVelocity.x - this.velocity.x) * accel;
        this.velocity.y += (targetVelocity.y - this.velocity.y) * accel;

        // Apply friction when no input
        if (magnitude === 0) {
            this.velocity.x *= Math.pow(this.FRICTION, deltaTime);
            this.velocity.y *= Math.pow(this.FRICTION, deltaTime);
        }

        // Update position
        this.position = {
            x: this.position.x + this.velocity.x * deltaTime,
            y: this.position.y + this.velocity.y * deltaTime
        };

        // Send position update to server if enough time has passed and position changed significantly
        this.maybeSendPositionUpdate();

        // Continue loop if there's movement or velocity
        const isMoving = this.activeDirections.size > 0 ||
                        Math.abs(this.velocity.x) > 0.001 ||
                        Math.abs(this.velocity.y) > 0.001;

        if (isMoving) {
            this.animationFrameId = requestAnimationFrame(this.updateMovement);
        } else {
            this.animationFrameId = null;
            // Send final position when stopped
            this.sendPositionUpdateNow();
        }
    };

    /**
     * Send position update to server if enough time has passed
     */
    private maybeSendPositionUpdate(): void {
        const now = performance.now();
        if (now - this.lastPositionSentTime < this.POSITION_SEND_INTERVAL) {
            return;
        }

        // Check if position changed significantly
        const distance = Math.abs(this.position.x - this.lastPositionSent.x) +
                        Math.abs(this.position.y - this.lastPositionSent.y);

        if (distance > 0.1) { // Send if moved more than 0.1 units
            this.sendPositionUpdateNow();
        }
    }

    /**
     * Send current position to server immediately
     */
    private sendPositionUpdateNow(): void {
        if (this.wsClient && this.wsClient.getState() === 'connected') {
            const roundedX = Math.round(this.position.x * 100) / 100;
            const roundedY = Math.round(this.position.y * 100) / 100;

            this.wsClient.sendPlayerUpdate(roundedX, roundedY);

            // Track what we sent to avoid conflicts with server updates
            this.lastSyncedPosition = { x: roundedX, y: roundedY };
            this.lastPositionSent = { x: roundedX, y: roundedY };
            this.lastPositionSentTime = performance.now();
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
        this.activeDirections.clear();
        this.stopMovementLoop();
    }

    /**
     * Update position from server game state
     * Only sync if this is the first time or if the server position differs significantly from what we sent
     */
    syncWithServer(player: PlayerOnBoard | null): void {
        if (!player) {
            this.playerId = null;
            this.position = { x: 0, y: 0 };
            this.lastSyncedPosition = { x: 0, y: 0 };
            this.isInitialized = false;
            return;
        }

        // First time initialization
        if (!this.isInitialized || this.playerId !== player.id) {
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
            this.position = { x: player.position.x, y: player.position.y };
            this.lastSyncedPosition = { x: player.position.x, y: player.position.y };
        }
    }

    /**
     * Stop the movement animation loop
     */
    private stopMovementLoop(): void {
        if (this.animationFrameId !== null) {
            cancelAnimationFrame(this.animationFrameId);
            this.animationFrameId = null;
        }
    }

    /**
     * Reset player state
     */
    reset(): void {
        this.stopMovementLoop();
        this.activeDirections.clear();
        this.position = { x: 0, y: 0 };
        this.playerId = null;
        this.pressedKeys.clear();
        this.wsClient = null;
        this.lastSyncedPosition = { x: 0, y: 0 };
        this.isInitialized = false;
        this.velocity = { x: 0, y: 0 };
        this.lastUpdateTime = 0;
        this.lastPositionSent = { x: 0, y: 0 };
        this.lastPositionSentTime = 0;
    }
}

export const playerState = new PlayerState();
