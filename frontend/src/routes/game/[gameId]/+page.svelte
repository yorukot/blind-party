<script lang="ts">
    import { PUBLIC_API_BASE_URL, PUBLIC_WS_BASE_URL } from '$env/static/public';
    import { gameState } from '$lib/api/game-state.svelte.js';
    import CountdownBar from '$lib/components/game/countdown-bar.svelte';
    import CountdownOverlay from '$lib/components/game/countdown-overlay.svelte';
    import GameBoardPanel from '$lib/components/game/game-board-panel.svelte';
    import PlayerMovementControls from '$lib/components/game/player-movement-controls.svelte';
    import PlayerRoster from '$lib/components/game/player-roster.svelte';
    import JoinForm from '$lib/components/ui/join-form.svelte';
    import { playerState } from '$lib/player-state.svelte.js';
    import type { PlayerOnBoard } from '$lib/types/player';
    import { onDestroy, onMount } from 'svelte';

    interface Props {
        params: {
            gameId: string;
        };
    }

    let { params }: Props = $props();

    // Connection state
    let connectionError = $state<string | null>(null);
    let isConnecting = $state(false);
    let username = $state('');
    let isJoined = $state(false);

    // Initialize game state
    gameState.initialize({
        apiBaseUrl: PUBLIC_API_BASE_URL || 'http://localhost:8080',
        wsBaseUrl: PUBLIC_WS_BASE_URL || 'ws://localhost:8080'
    });

    // Get reactive data from game state
    let mapSize = $derived(gameState.mapSize);
    let gameMap = $derived(gameState.gameMap);
    let players = $derived(gameState.players);
    let localPlayer = $derived(gameState.localPlayer);
    let connectionState = $derived(gameState.connectionState);


    // Join game function
    async function joinGame() {
        if (!username.trim()) return;

        isConnecting = true;
        connectionError = null;

        try {
            await gameState.joinGame(params.gameId, username.trim());
            isJoined = true;
        } catch (error) {
            connectionError = error instanceof Error ? error.message : 'Failed to join game';
        } finally {
            isConnecting = false;
        }
    }

    // Sync player state with game state
    $effect(() => {
        if (localPlayer) {
            playerState.syncWithGameState(localPlayer);
        }
    });

    // Set up position update callback
    $effect(() => {
        if (isJoined && connectionState === 'connected') {
            playerState.setPositionUpdateCallback((x, y) => {
                gameState.updatePlayerPosition(x, y);
            });
        } else {
            playerState.setPositionUpdateCallback(null);
        }
    });

    let selfPlayerOnBoard = $derived.by(() => {
        const player = playerState.localPlayer;
        if (!player) return null;
        return {
            ...player,
            position: playerState.localPosition
        };
    });
    let selfPlayerSummary = $derived(localPlayer);

    let otherPlayersOnBoard = $derived.by(() =>
        players
            .filter((player) => player.position && player.id !== localPlayer?.id)
            .map(
                (player) =>
                    ({
                        ...player,
                        position: player.position!
                    }) as PlayerOnBoard
            )
    );

    const MAX_PLAYER_SPEED = 4; // tiles per second
    const ACCELERATION_RATE = 12; // tiles per second squared
    const FRICTION_RATE = 10; // tiles per second squared

    function clampToBoard(value: number) {
        const maxIndex = Math.max(0, mapSize - 1);
        if (value < 0) {
            return 0;
        }
        if (value > maxIndex) {
            return maxIndex;
        }
        return value;
    }

    function updateLocalPlayer(deltaSeconds: number) {
        if (!playerState.localPlayerId) {
            return;
        }

        const directions = playerState.activeDirections;
        const hasInput = directions.size > 0;

        const velocity = playerState.localVelocity;
        let nextVx = velocity.x;
        let nextVy = velocity.y;
        const currentPosition = playerState.localPosition;

        if (hasInput) {
            let dx = 0;
            let dy = 0;

            if (directions.has('up')) {
                dy -= 1;
            }
            if (directions.has('down')) {
                dy += 1;
            }
            if (directions.has('left')) {
                dx -= 1;
            }
            if (directions.has('right')) {
                dx += 1;
            }

            const magnitude = Math.hypot(dx, dy);
            if (magnitude > 0) {
                const ax = (dx / magnitude) * ACCELERATION_RATE;
                const ay = (dy / magnitude) * ACCELERATION_RATE;
                nextVx += ax * deltaSeconds;
                nextVy += ay * deltaSeconds;
            }
        } else {
            const speed = Math.hypot(nextVx, nextVy);
            if (speed > 0) {
                const decel = Math.min(speed, FRICTION_RATE * deltaSeconds);
                const scale = (speed - decel) / speed;
                nextVx *= scale;
                nextVy *= scale;
            }
        }

        const nextSpeed = Math.hypot(nextVx, nextVy);
        if (nextSpeed > MAX_PLAYER_SPEED) {
            const scale = MAX_PLAYER_SPEED / nextSpeed;
            nextVx *= scale;
            nextVy *= scale;
        }

        let nextX = currentPosition.x + nextVx * deltaSeconds;
        let nextY = currentPosition.y + nextVy * deltaSeconds;

        const clampedX = clampToBoard(nextX);
        const clampedY = clampToBoard(nextY);

        if (clampedX !== nextX) {
            nextX = clampedX;
            nextVx = 0;
        }
        if (clampedY !== nextY) {
            nextY = clampedY;
            nextVy = 0;
        }

        playerState.updateLocalPlayerVelocity({ x: nextVx, y: nextVy });
        playerState.updateLocalPlayerPosition({ x: nextX, y: nextY });
    }

    // Get remaining time from game state
    let remainingSeconds = $derived(Math.ceil(gameState.remainingTime));

    onMount(() => {
        let rafId = 0;
        let lastTimestamp = 0;
        const loop = (timestamp: number) => {
            if (!lastTimestamp) {
                lastTimestamp = timestamp;
            }

            const deltaSeconds = (timestamp - lastTimestamp) / 1000;
            lastTimestamp = timestamp;

            updateLocalPlayer(deltaSeconds);
            rafId = requestAnimationFrame(loop);
        };

        rafId = requestAnimationFrame(loop);

        return () => {
            cancelAnimationFrame(rafId);
        };
    });

    onDestroy(() => {
        // Clean up when component is destroyed
        gameState.disconnect();
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

            <CountdownBar duration={90} fillColor="#facc15" />

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
