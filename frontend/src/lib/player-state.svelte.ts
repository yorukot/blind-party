import { SvelteSet } from 'svelte/reactivity';

export type Direction = 'up' | 'down' | 'left' | 'right';

/*
Class storing our own player's state
 **/
class PlayerState {
    activeDirections = $state<SvelteSet<Direction>>(new SvelteSet());
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
}

export const playerState = new PlayerState();
