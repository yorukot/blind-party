<script lang="ts">
    import { PUBLIC_API_BASE_URL, PUBLIC_WS_BASE_URL } from '$env/static/public';
    import { createWebSocketClient, type WebSocketGameClient } from '$lib/api/websocket';
    import CountdownOverlay from '$lib/components/game/countdown-overlay.svelte';
    import GameBoardPanel from '$lib/components/game/game-board-panel.svelte';
    import GameStatusBar from '$lib/components/game/game-status-bar.svelte';
    import PlayerMovementControls from '$lib/components/game/player-movement-controls.svelte';
    import PlayerRoster from '$lib/components/game/player-roster.svelte';
    import JoinForm from '$lib/components/ui/join-form.svelte';
    import { playerState } from '$lib/player-state.svelte.js';
    import type { GameStateResponse } from '$lib/types/game';
    import { gamePlayerToPlayerSummary } from '$lib/types/game';
    import type { PlayerOnBoard, PlayerSummary } from '$lib/types/player';
    import { onDestroy } from 'svelte';

    interface Props {
        params: {
            gameId: string;
        };
    }

    let { params }: Props = $props();

    let isJoined = $state(false);
    let username = $state('');
    let isConnecting = $state(false);
    let connectionError = $state<string | null>(null);
    let connectionState = $state('disconnected');

    // WebSocket client
    let wsClient: WebSocketGameClient | null = null;

    // Game state from WebSocket
    let gameState = $state<GameStateResponse>({
        game_id: '',
        phase: 'pre-game',
        players: [],
        map: [],
        countdown_seconds: null
    });

    // Derived values for UI
    let players = $derived(gameState.players.map(gamePlayerToPlayerSummary));
    let mapSize = $derived(Math.max(gameState.map.length || 10, gameState.map[0]?.length || 10));
    let gameMap = $derived(gameState.map);
    let remainingSeconds = $derived(gameState.countdown_seconds || 0);

    // Player-specific derived values
    let selfPlayerSummary = $derived(players.find((p) => p.name === username) || null);

    // Use local player state position for immediate movement feedback
    let selfPlayerOnBoard = $derived.by(() => {
        if (!selfPlayerSummary || !selfPlayerSummary.position) {
            return null;
        }

        // Create player object with local position from playerState
        return {
            ...selfPlayerSummary,
            position: playerState.position
        } as PlayerOnBoard;
    });

    let otherPlayersOnBoard = $derived(
        players.filter((p) => p.name !== username && p.position) as PlayerOnBoard[]
    );

    function createWebSocketGameClient() {
        return createWebSocketClient({
            autoReconnect: true,
            maxReconnectAttempts: 5,
            reconnectDelay: 1000,
            pingInterval: 30000
        });
    }

    function setupEventHandlers(client: WebSocketGameClient) {
        client.on('onStateChange', (state) => {
            connectionState = state;
            if (state === 'connected') {
                isJoined = true;
            }
        });

        client.on('onGameUpdate', (updatedGameState) => {
            gameState = updatedGameState;

            // Sync player position with server
            const selfPlayerSummary = updatedGameState.players
                .map(gamePlayerToPlayerSummary)
                .find(p => p.name === username);

            const selfPlayerOnBoard = selfPlayerSummary && selfPlayerSummary.position
                ? (selfPlayerSummary as PlayerOnBoard)
                : null;

            if (selfPlayerOnBoard) {
                playerState.syncWithServer(selfPlayerOnBoard);
            }
        });

        client.on('onError', (error) => {
            connectionError = error;
            isConnecting = false;
        });
    }

    async function joinGame() {
        if (!username.trim()) {
            connectionError = 'Please enter a username';
            return;
        }

        isConnecting = true;
        connectionError = null;

        try {
            // Initialize WebSocket client
            wsClient = createWebSocketGameClient();

            // Set up event listeners
            setupEventHandlers(wsClient);

            // Connect WebSocket client to player state
            playerState.setWebSocketClient(wsClient);

            // Connect to the game
            await wsClient.connect(params.gameId, username.trim());

            console.log('[Game] WebSocket connected, initializing player state');
        } catch (error) {
            connectionError = error instanceof Error ? error.message : 'Failed to connect to game';
            isConnecting = false;
        } finally {
            isConnecting = false;
        }
    }

    // Cleanup on component destroy
    onDestroy(() => {
        if (wsClient) {
            wsClient.disconnect();
        }
        playerState.reset();
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 text-white">
    {#if !isJoined}
        <JoinForm
            {params}
            bind:username
            {isConnecting}
            {connectionError}
            {connectionState}
            onJoin={joinGame}
        />
    {:else}
        <!-- Game interface -->
        <div class="mx-auto flex max-w-6xl flex-col gap-10 px-4 py-10 sm:px-6 sm:py-12">
            <header class="flex flex-col gap-3 text-center lg:text-left">
                <p class="text-sm tracking-[0.35em] text-blue-200/80 uppercase">
                    Blind Party - {username}
                </p>
                <h1
                    class="font-minecraft text-3xl tracking-wider text-yellow-300 uppercase drop-shadow-[4px_4px_0px_rgba(0,0,0,0.65)] sm:text-4xl"
                >
                    Game ID: <span class="text-white">{params.gameId}</span>
                </h1>
                <p class="text-base text-blue-100/80">
                    {#if connectionState === 'connected'}
                        Connected to game server - {gameState.phase}
                    {:else}
                        Connection status: {connectionState}
                    {/if}
                </p>
            </header>

            {#if gameState.phase === 'pre-game'}
                <GameStatusBar
                    label="Waiting for Players"
                    displayText="{players.length}/2"
                    progress={(players.length / 2) * 100}
                    fillColor="#f59e0b"
                />
            {:else if gameState.phase === 'in-game'}
                <GameStatusBar
                    label="Game In Progress"
                    displayText={remainingSeconds > 0 ? `${remainingSeconds}s` : 'Active'}
                    progress={remainingSeconds > 0 ? (remainingSeconds / 30) * 100 : 100}
                    fillColor="#3b82f6"
                />
            {:else if gameState.phase === 'settlement'}
                <GameStatusBar
                    label="Game Ended"
                    displayText="Final Results"
                    progress={100}
                    fillColor="#10b981"
                />
            {:else}
                <GameStatusBar
                    label="Game Status"
                    displayText="Ready"
                    progress={100}
                    fillColor="#22c55e"
                />
            {/if}

            {#if remainingSeconds > 0}
                <CountdownOverlay {remainingSeconds} />
            {/if}

            <div class="flex flex-col gap-8 lg:flex-row">
                <GameBoardPanel
                    {mapSize}
                    {gameMap}
                    players={otherPlayersOnBoard}
                    selfPlayer={selfPlayerOnBoard}
                />
                <!-- Show movement controls between board and roster on mobile -->
                <div class="lg:hidden">
                    <PlayerMovementControls />
                </div>
                <PlayerRoster {players} selfPlayer={selfPlayerSummary} />
            </div>

            <!-- Keep the original controls visible only on large screens (desktop) -->
            <div class="hidden lg:block">
                <PlayerMovementControls />
            </div>
        </div>
    {/if}
</div>
