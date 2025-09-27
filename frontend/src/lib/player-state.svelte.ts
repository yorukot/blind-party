import type { PlayerOnBoard, PlayerPosition } from '$lib/types/player';
import { SvelteSet } from 'svelte/reactivity';

export type Direction = 'up' | 'down' | 'left' | 'right';

/**
 * Callback function type for position updates.
 */
export type PositionUpdateCallback = (x: number, y: number) => void;

/**
 * Class storing our own player's state.
 */
class PlayerState {
    activeDirections = $state<SvelteSet<Direction>>(new SvelteSet());
    localPlayer = $state<PlayerOnBoard | null>(null);
    localPlayerId = $state<string | null>(null);
    localVelocity = $state<{ x: number; y: number }>({ x: 0, y: 0 });
    localPosition = $state<PlayerPosition>({ x: 0, y: 0 });
    private pressedKeys = new Set<string>();
    private positionUpdateCallback: PositionUpdateCallback | null = null;
    private lastSentPosition: PlayerPosition | null = null;
    private lastSentTime: number = 0;
    private readonly POSITION_UPDATE_RATE_MS = 50; // 20Hz = 1000ms / 20 = 50ms
    private readonly POSITION_EPSILON = 0.005;
    private currentLocalPlayerSnapshot: PlayerOnBoard | null = null;
    private currentLocalPosition: PlayerPosition = { x: 0, y: 0 };

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
        const snapshot = player ? this.clonePlayerSnapshot(player) : null;
        this.localPlayer = snapshot;
        this.currentLocalPlayerSnapshot = snapshot;
        this.localPlayerId = player?.id ?? null;
        this.localVelocity = { x: 0, y: 0 };

        if (snapshot) {
            const positionSnapshot = this.clonePosition(snapshot.position);
            this.localPosition = positionSnapshot;
            this.currentLocalPosition = this.clonePosition(positionSnapshot);
            this.lastSentPosition = this.clonePosition(positionSnapshot);
            this.lastSentTime = 0;
        } else {
            this.localPosition = { x: 0, y: 0 };
            this.currentLocalPosition = { x: 0, y: 0 };
            this.lastSentPosition = null;
            this.lastSentTime = 0;
        }
    }

    setLocalPlayerId(id: string | null) {
        this.localPlayerId = id;
        if (!id || this.currentLocalPlayerSnapshot?.id !== id) {
            this.localPlayer = null;
            this.currentLocalPlayerSnapshot = null;
            this.localVelocity = { x: 0, y: 0 };
            this.localPosition = { x: 0, y: 0 };
            this.currentLocalPosition = { x: 0, y: 0 };
            this.lastSentPosition = null;
            this.lastSentTime = 0;
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
     */
    setPositionUpdateCallback(callback: PositionUpdateCallback | null): void {
        this.positionUpdateCallback = callback;
    }

    /**
     * Update the local player position.
     * This method also triggers the position update callback if set.
     */
    updateLocalPlayerPosition(position: PlayerPosition, skipUpdate = false): void {
        if (!this.currentLocalPlayerSnapshot) {
            return;
        }

        const nextPosition = this.roundToPosition(position);
        const currentPosition = this.currentLocalPosition;

        if (this.arePositionsClose(currentPosition, nextPosition)) {
            return;
        }

        this.localPosition = nextPosition;
        this.currentLocalPosition = this.clonePosition(nextPosition);

        // Send position update if callback is set and we're not skipping
        if (
            !skipUpdate &&
            this.positionUpdateCallback &&
            this.shouldSendPositionUpdate(nextPosition)
        ) {
            this.positionUpdateCallback(nextPosition.x, nextPosition.y);
            this.lastSentPosition = this.clonePosition(nextPosition);
            this.lastSentTime = Date.now();
        }
    }

    /**
     * Determine if we should send a position update.
     * This reduces unnecessary callbacks by only sending when position changes significantly
     * and respects the 20Hz rate limit.
     */
    private shouldSendPositionUpdate(newPosition: PlayerPosition): boolean {
        const now = Date.now();

        // Rate limiting: don't send more than 20Hz (every 50ms)
        if (now - this.lastSentTime < this.POSITION_UPDATE_RATE_MS) {
            return false;
        }

        if (!this.lastSentPosition) {
            return true;
        }

        return !this.arePositionsClose(this.lastSentPosition, newPosition, 0.01);
    }

    /**
     * Sync the local player with data from the game state.
     * This is called when we receive updates from the server.
     * Only syncs position if the server position is significantly different from client position.
     */
    syncWithGameState(player: PlayerOnBoard | null): void {
        if (!player) {
            this.localPlayer = null;
            this.currentLocalPlayerSnapshot = null;
            this.localPlayerId = null;
            this.localVelocity = { x: 0, y: 0 };
            this.localPosition = { x: 0, y: 0 };
            this.currentLocalPosition = { x: 0, y: 0 };
            this.lastSentPosition = null;
            return;
        }

        const currentSnapshot = this.currentLocalPlayerSnapshot;
        const snapshot = this.clonePlayerSnapshot(player);

        if (!currentSnapshot || currentSnapshot.id !== player.id) {
            const initialPosition = this.clonePosition(snapshot.position);
            this.localPlayer = snapshot;
            this.currentLocalPlayerSnapshot = snapshot;
            this.localPlayerId = player.id;
            this.localPosition = initialPosition;
            this.currentLocalPosition = this.clonePosition(initialPosition);
            this.localVelocity = { x: 0, y: 0 };
            this.lastSentPosition = this.clonePosition(initialPosition);
            this.lastSentTime = 0;
            return;
        }

        this.localPlayer = snapshot;
        this.currentLocalPlayerSnapshot = snapshot;
        this.localPlayerId = player.id;

        const currentPosition = this.currentLocalPosition;
        const SYNC_THRESHOLD = 0.5; // Only sync if positions differ by more than 0.5 tiles
        const distance = this.calculateDistance(currentPosition, snapshot.position);

        if (distance > SYNC_THRESHOLD) {
            const resynced = this.clonePosition(snapshot.position);
            this.localPosition = resynced;
            this.currentLocalPosition = this.clonePosition(resynced);
            this.localVelocity = { x: 0, y: 0 };
            this.lastSentPosition = this.clonePosition(resynced);
            this.lastSentTime = 0;
        }
    }

    /**
     * Reset the player state to initial values.
     */
    reset(): void {
        this.activeDirections.clear();
        this.localPlayer = null;
        this.currentLocalPlayerSnapshot = null;
        this.localPlayerId = null;
        this.localVelocity = { x: 0, y: 0 };
        this.localPosition = { x: 0, y: 0 };
        this.currentLocalPosition = { x: 0, y: 0 };
        this.pressedKeys.clear();
        this.lastSentPosition = null;
        this.lastSentTime = 0;
        this.positionUpdateCallback = null;
    }

    private clonePlayerSnapshot(player: PlayerOnBoard): PlayerOnBoard {
        return {
            ...player,
            position: this.clonePosition(player.position)
        };
    }

    private roundToPosition(position: PlayerPosition): PlayerPosition {
        return {
            x: this.roundToGridCoordinate(position.x),
            y: this.roundToGridCoordinate(position.y)
        };
    }

    private arePositionsClose(
        a: PlayerPosition | null,
        b: PlayerPosition,
        epsilon: number = this.POSITION_EPSILON
    ): boolean {
        if (!a) {
            return false;
        }

        return Math.abs(a.x - b.x) < epsilon && Math.abs(a.y - b.y) < epsilon;
    }

    private clonePosition(position: PlayerPosition): PlayerPosition {
        return { x: position.x, y: position.y };
    }

    private calculateDistance(a: PlayerPosition, b: PlayerPosition): number {
        const dx = b.x - a.x;
        const dy = b.y - a.y;
        return Math.sqrt(dx * dx + dy * dy);
    }
}

export const playerState = new PlayerState();
