export type PlayerStatus = 'spectating' | 'ingame' | 'eliminated';

export interface PlayerPosition {
    /**
     * Zero-based column index within the game board grid.
     */
    x: number;
    /**
     * Zero-based row index within the game board grid.
     */
    y: number;
}

export interface PlayerSummary {
    id: string;
    name: string;
    status: PlayerStatus;
    accent: string;
    /**
     * Optional grid coordinate for rendering on the game board.
     */
    position?: PlayerPosition;
}

export interface PlayerOnBoard extends PlayerSummary {
    position: PlayerPosition;
}
