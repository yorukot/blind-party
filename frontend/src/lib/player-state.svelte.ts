import { SvelteSet } from 'svelte/reactivity';
import type { PlayerOnBoard, PlayerPosition } from '$lib/types/player';

export type Direction = 'up' | 'down' | 'left' | 'right';

/*
Class storing our own player's state
 **/
class PlayerState {
    activeDirections = $state<SvelteSet<Direction>>(new SvelteSet());
    localPlayer = $state<PlayerOnBoard | null>(null);
    localPlayerId = $state<string | null>(null);
    localVelocity = $state<{ x: number; y: number }>({ x: 0, y: 0 });
    private pressedKeys = new Set<string>();

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

    updateLocalPlayerPosition(position: PlayerPosition) {
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

        this.localPlayer = {
            ...current,
            position: { x: roundedX, y: roundedY }
        };
    }

    updateLocalPlayerVelocity(velocity: { x: number; y: number }) {
        const roundedX = this.roundToGridCoordinate(velocity.x);
        const roundedY = this.roundToGridCoordinate(velocity.y);
        this.localVelocity = { x: roundedX, y: roundedY };
    }
}

export const playerState = new PlayerState();
