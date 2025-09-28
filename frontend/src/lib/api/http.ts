import { PUBLIC_API_BASE_URL } from '$env/static/public';

/**
 * Response type for creating a new game
 */
export interface CreateGameResponse {
    game_id: string;
}

/**
 * Base API error interface
 */
export interface ApiError {
    message: string;
    status: number;
    details?: unknown;
}

/**
 * Custom error class for API errors
 */
export class HttpApiError extends Error {
    constructor(
        public status: number,
        message: string,
        public details?: unknown
    ) {
        super(message);
        this.name = 'HttpApiError';
    }
}

/**
 * HTTP API client configuration
 */
class HttpApiClient {
    private readonly baseUrl: string;

    constructor() {
        this.baseUrl = PUBLIC_API_BASE_URL;
        if (!this.baseUrl) {
            throw new Error('PUBLIC_API_BASE_URL environment variable is not set');
        }
    }

    /**
     * Make a HTTP request with proper error handling
     */
    private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;

        try {
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });

            if (!response.ok) {
                const errorData = await response.text();
                let errorMessage = `HTTP ${response.status}: ${response.statusText}`;

                try {
                    const parsedError = JSON.parse(errorData);
                    errorMessage = parsedError.message || errorMessage;
                } catch {
                    // If parsing fails, use the raw text or default message
                    errorMessage = errorData || errorMessage;
                }

                throw new HttpApiError(response.status, errorMessage, errorData);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }

            // If response is not JSON, return empty object
            return {} as T;
        } catch (error) {
            if (error instanceof HttpApiError) {
                throw error;
            }

            // Network or other errors
            throw new HttpApiError(
                0,
                `Network error: ${error instanceof Error ? error.message : 'Unknown error'}`,
                error
            );
        }
    }

    /**
     * Creates a new game instance
     *
     * @returns Promise resolving to the game creation response
     * @throws HttpApiError if the request fails
     */
    async createGame(): Promise<CreateGameResponse> {
        return this.request<CreateGameResponse>('/api/game/', {
            method: 'POST'
        });
    }

    /**
     * Get the base URL being used by the client
     */
    getBaseUrl(): string {
        return this.baseUrl;
    }
}

/**
 * Singleton HTTP API client instance
 */
export const httpApi = new HttpApiClient();

/**
 * Convenience function to create a new game
 */
export async function createGame(): Promise<CreateGameResponse> {
    return httpApi.createGame();
}
