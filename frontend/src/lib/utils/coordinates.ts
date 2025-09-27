import type { PlayerPosition } from '$lib/types/player';

/**
 * Coordinate conversion utilities between frontend grid coordinates and API coordinates.
 *
 * Frontend: 0-based integer grid coordinates (0-19 for a 20x20 map)
 * API: 1.0-21.0 floating-point coordinates with block centers at .5 positions
 *
 * Examples:
 * - Frontend (0, 0) -> API (1.5, 1.5) - center of top-left block
 * - Frontend (19, 19) -> API (20.5, 20.5) - center of bottom-right block
 */

export const COORDINATE_CONSTANTS = {
    API_MIN: 1.0,
    API_MAX: 21.0,
    GRID_SIZE: 20,
    BLOCK_CENTER_OFFSET: 0.5
} as const;

/**
 * Convert frontend grid coordinates to API coordinates.
 * Frontend grid coordinates are 0-based integers.
 * API coordinates place block centers at .5 positions (1.5, 2.5, etc.)
 */
export function gridToApi(gridPosition: PlayerPosition): { pos_x: number; pos_y: number } {
    return {
        pos_x: Math.round((gridPosition.x + COORDINATE_CONSTANTS.BLOCK_CENTER_OFFSET + COORDINATE_CONSTANTS.API_MIN) * 100) / 100,
        pos_y: Math.round((gridPosition.y + COORDINATE_CONSTANTS.BLOCK_CENTER_OFFSET + COORDINATE_CONSTANTS.API_MIN) * 100) / 100
    };
}

/**
 * Convert API coordinates to frontend grid coordinates.
 * Rounds to the nearest block center and converts to 0-based grid position.
 */
export function apiToGrid(apiPosition: { pos_x: number; pos_y: number }): PlayerPosition {
    return {
        x: Math.round(apiPosition.pos_x - COORDINATE_CONSTANTS.API_MIN - COORDINATE_CONSTANTS.BLOCK_CENTER_OFFSET),
        y: Math.round(apiPosition.pos_y - COORDINATE_CONSTANTS.API_MIN - COORDINATE_CONSTANTS.BLOCK_CENTER_OFFSET)
    };
}

/**
 * Validate that API coordinates are within the valid game bounds.
 */
export function validateApiCoordinates(position: { pos_x: number; pos_y: number }): boolean {
    return (
        position.pos_x >= COORDINATE_CONSTANTS.API_MIN &&
        position.pos_x <= COORDINATE_CONSTANTS.API_MAX &&
        position.pos_y >= COORDINATE_CONSTANTS.API_MIN &&
        position.pos_y <= COORDINATE_CONSTANTS.API_MAX
    );
}

/**
 * Validate that grid coordinates are within the valid game bounds.
 */
export function validateGridCoordinates(position: PlayerPosition): boolean {
    return (
        position.x >= 0 &&
        position.x < COORDINATE_CONSTANTS.GRID_SIZE &&
        position.y >= 0 &&
        position.y < COORDINATE_CONSTANTS.GRID_SIZE
    );
}

/**
 * Clamp API coordinates to valid bounds.
 */
export function clampApiCoordinates(position: { pos_x: number; pos_y: number }): { pos_x: number; pos_y: number } {
    return {
        pos_x: Math.max(COORDINATE_CONSTANTS.API_MIN, Math.min(COORDINATE_CONSTANTS.API_MAX, position.pos_x)),
        pos_y: Math.max(COORDINATE_CONSTANTS.API_MIN, Math.min(COORDINATE_CONSTANTS.API_MAX, position.pos_y))
    };
}

/**
 * Clamp grid coordinates to valid bounds.
 */
export function clampGridCoordinates(position: PlayerPosition): PlayerPosition {
    return {
        x: Math.max(0, Math.min(COORDINATE_CONSTANTS.GRID_SIZE - 1, position.x)),
        y: Math.max(0, Math.min(COORDINATE_CONSTANTS.GRID_SIZE - 1, position.y))
    };
}

/**
 * Round API coordinates to the maximum precision allowed by the API (2 decimal places).
 */
export function roundApiCoordinates(position: { pos_x: number; pos_y: number }): { pos_x: number; pos_y: number } {
    return {
        pos_x: Math.round(position.pos_x * 100) / 100,
        pos_y: Math.round(position.pos_y * 100) / 100
    };
}

/**
 * Calculate the distance between two API coordinate positions.
 */
export function calculateApiDistance(
    pos1: { pos_x: number; pos_y: number },
    pos2: { pos_x: number; pos_y: number }
): number {
    const dx = pos2.pos_x - pos1.pos_x;
    const dy = pos2.pos_y - pos1.pos_y;
    return Math.sqrt(dx * dx + dy * dy);
}

/**
 * Calculate the distance between two grid coordinate positions.
 */
export function calculateGridDistance(pos1: PlayerPosition, pos2: PlayerPosition): number {
    const dx = pos2.x - pos1.x;
    const dy = pos2.y - pos1.y;
    return Math.sqrt(dx * dx + dy * dy);
}