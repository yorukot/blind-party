import type { CreateGameResponse, APIClientConfig } from '$lib/types/api';

/**
 * HTTP API client for the Color Rush Survival game backend.
 * Handles REST API endpoints for game management.
 */
export class HTTPClient {
    private baseUrl: string;

    constructor(config: Pick<APIClientConfig, 'apiBaseUrl'>) {
        this.baseUrl = config.apiBaseUrl.endsWith('/')
            ? config.apiBaseUrl.slice(0, -1)
            : config.apiBaseUrl;
    }

    /**
     * Create a new game instance.
     * @returns Promise with the game ID
     */
    async createGame(): Promise<CreateGameResponse> {
        const response = await fetch(`${this.baseUrl}/api/game/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            throw new HTTPError(
                `Failed to create game: ${response.status} ${response.statusText}`,
                response.status,
                response.statusText
            );
        }

        const data = await response.json();

        // Validate response structure
        if (!data || typeof data.game_id !== 'string') {
            throw new HTTPError(
                'Invalid response format: missing or invalid game_id',
                response.status,
                'Invalid Response'
            );
        }

        return data as CreateGameResponse;
    }

    /**
     * Check if the API server is reachable.
     * This is a utility method for connection testing.
     */
    async healthCheck(): Promise<boolean> {
        try {
            const response = await fetch(`${this.baseUrl}/api/health`, {
                method: 'GET',
                signal: AbortSignal.timeout(5000), // 5 second timeout
            });
            return response.ok;
        } catch {
            return false;
        }
    }
}

/**
 * Custom error class for HTTP API errors.
 */
export class HTTPError extends Error {
    constructor(
        message: string,
        public readonly status: number,
        public readonly statusText: string
    ) {
        super(message);
        this.name = 'HTTPError';
    }
}

/**
 * Default configuration values for the HTTP client.
 */
export const DEFAULT_HTTP_CONFIG = {
    apiBaseUrl: 'http://localhost:8080', // Default backend URL
} as const;

/**
 * Create an HTTP client instance with default configuration.
 * Can be overridden with environment variables or custom config.
 */
export function createHTTPClient(customConfig?: Partial<APIClientConfig>): HTTPClient {
    const config = {
        apiBaseUrl: customConfig?.apiBaseUrl ?? DEFAULT_HTTP_CONFIG.apiBaseUrl,
    };

    return new HTTPClient(config);
}