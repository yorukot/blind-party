import { SvelteSet } from 'svelte/reactivity';
import type { PlayerOnBoard, PlayerPosition } from '$lib/types/player';

export type Direction = 'up' | 'down' | 'left' | 'right';

/**
 * Callback function type for position updates.
 */
export type PositionUpdateCallback = (x: number, y: number) => void;

/**
 * Class storing our own player's state with API integration.
 */
class PlayerState {
    activeDirections = $state<SvelteSet<Direction>>(new SvelteSet());
    localPlayer = $state<PlayerOnBoard | null>(null);
    localPlayerId = $state<string | null>(null);
    localVelocity = $state<{ x: number; y: number }>({ x: 0, y: 0 });
    private pressedKeys = new Set<string>();
    private positionUpdateCallback: PositionUpdateCallback | null = null;
    private lastSentPosition: PlayerPosition | null = null;

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

    triggerMove(direction: Direction) {
        this.activeDirections.add(direction);
    }

    clearDirection(direction: Direction) {
        this.activeDirections.delete(direction);
    }

    handleKeyDown(event: KeyboardEvent) {
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

    handleKeyUp(event: KeyboardEvent) {
        const direction = this.getDirectionFromKey(event.key);
        this.pressedKeys.delete(event.key);
        if (!direction) {
            return;
        }

        this.clearDirection(direction);
    }

    clearPressedKeys() {
        this.pressedKeys.clear();
    }

    setLocalPlayer(player: PlayerOnBoard | null) {
        this.localPlayer = player;
        this.localPlayerId = player?.id ?? null;
        this.localVelocity = { x: 0, y: 0 };
    }

    setLocalPlayerId(id: string | null) {
        this.localPlayerId = id;
        if (!id || this.localPlayer?.id !== id) {
            this.localPlayer = null;
            this.localVelocity = { x: 0, y: 0 };
        }
    }

    private roundToGridCoordinate(value: number) {
        return Math.round(value * 100) / 100;
    }


    updateLocalPlayerVelocity(velocity: { x: number; y: number }) {
        const roundedX = this.roundToGridCoordinate(velocity.x);
        const roundedY = this.roundToGridCoordinate(velocity.y);
        this.localVelocity = { x: roundedX, y: roundedY };
    }

    /**
     * Set a callback function to be called when the player position changes.
     * This is used to send position updates to the API.
     */
    setPositionUpdateCallback(callback: PositionUpdateCallback | null): void {
        this.positionUpdateCallback = callback;
    }

    /**
     * Update the local player position and optionally send to API.
     * This method also triggers the position update callback if set.
     */
    updateLocalPlayerPosition(position: PlayerPosition, skipAPIUpdate = false): void {
        const current = this.localPlayer;
        if (!current) {
            return;
        }

        const roundedX = this.roundToGridCoordinate(position.x);
        const roundedY = this.roundToGridCoordinate(position.y);
        const epsilon = 0.0001;

        if (
            Math.abs(roundedX - current.position.x) < epsilon &&
            Math.abs(roundedY - current.position.y) < epsilon
        ) {
            return;
        }

        const newPosition = { x: roundedX, y: roundedY };

        this.localPlayer = {
            ...current,
            position: newPosition
        };

        // Send position update to API if callback is set and we're not skipping
        if (!skipAPIUpdate && this.positionUpdateCallback && this.shouldSendPositionUpdate(newPosition)) {
            this.positionUpdateCallback(roundedX, roundedY);
            this.lastSentPosition = newPosition;
        }
    }

    /**
     * Determine if we should send a position update to the API.
     * This reduces unnecessary network traffic by only sending when position changes significantly.
     */
    private shouldSendPositionUpdate(newPosition: PlayerPosition): boolean {
        if (!this.lastSentPosition) {
            return true;
        }

        const dx = Math.abs(newPosition.x - this.lastSentPosition.x);
        const dy = Math.abs(newPosition.y - this.lastSentPosition.y);
        const threshold = 0.01; // Minimum movement threshold

        return dx >= threshold || dy >= threshold;
    }

    /**
     * Sync the local player with data from the game state.
     * This is called when we receive updates from the API.
     */
    syncWithGameState(player: PlayerOnBoard | null): void {
        if (!player) {
            this.localPlayer = null;
            this.localPlayerId = null;
            this.localVelocity = { x: 0, y: 0 };
            this.lastSentPosition = null;
            return;
        }

        // Update without triggering API update to avoid feedback loop
        this.localPlayer = player;
        this.localPlayerId = player.id;
        this.lastSentPosition = player.position;
    }

    /**
     * Reset the player state to initial values.
     */
    reset(): void {
        this.activeDirections.clear();
        this.localPlayer = null;
        this.localPlayerId = null;
        this.localVelocity = { x: 0, y: 0 };
        this.pressedKeys.clear();
        this.lastSentPosition = null;
        this.positionUpdateCallback = null;
    }
}

export const playerState = new PlayerState();
