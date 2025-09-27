<script lang="ts">
    import CountdownBar from '$lib/components/game/countdown-bar.svelte';
    import CountdownOverlay from '$lib/components/game/countdown-overlay.svelte';
    import GameBoardPanel from '$lib/components/game/game-board-panel.svelte';
    import PlayerMovementControls from '$lib/components/game/player-movement-controls.svelte';
    import PlayerRoster from '$lib/components/game/player-roster.svelte';
    import { playerState } from '$lib/player-state.svelte.js';
    import type { PlayerOnBoard, PlayerSummary } from '$lib/types/player';
    import { onMount } from 'svelte';

    interface Props {
        params: {
            gameId: string;
        };
    }

    let { params }: Props = $props();
    let mapSize = $state(18);
    let players = $state<PlayerSummary[]>([
        {
            id: '1',
            name: 'PixelPanda',
            status: 'ingame',
            accent: 'from-emerald-400 to-emerald-600',
            position: { x: 4, y: 6 }
        },
        {
            id: '2',
            name: 'ShadowFox',
            status: 'ingame',
            accent: 'from-blue-400 to-indigo-600',
            position: { x: 10, y: 8 }
        },
        {
            id: '4',
            name: 'ByteKnight',
            status: 'eliminated',
            accent: 'from-slate-500 to-slate-700',
            position: { x: 14, y: 2 }
        }
    ]);

    const sampleLocalPlayer: PlayerOnBoard = {
        id: '3',
        name: 'NeonNova',
        status: 'spectating',
        accent: 'from-pink-400 to-rose-600',
        position: { x: 2, y: 12 }
    };

    if (!playerState.localPlayer) {
        playerState.setLocalPlayer(sampleLocalPlayer);
    }

    let selfPlayerOnBoard = $derived.by(() => playerState.localPlayer);
    let selfPlayerSummary = $derived.by(() => playerState.localPlayer);

    let otherPlayersOnBoard = $derived.by(() =>
        players
            .filter((player) => player.position)
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
        const current = playerState.localPlayer;
        if (!current) {
            return;
        }

        const directions = playerState.activeDirections;
        const hasInput = directions.size > 0;

        const velocity = playerState.localVelocity;
        let nextVx = velocity.x;
        let nextVy = velocity.y;

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

        let nextX = current.position.x + nextVx * deltaSeconds;
        let nextY = current.position.y + nextVy * deltaSeconds;

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

    let remainingSeconds = $state(5);
    onMount(() => {
        setInterval(() => {
            if (remainingSeconds > 0) {
                remainingSeconds -= 1;
            }
        }, 1000);
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 text-white">
    <div class="mx-auto flex max-w-6xl flex-col gap-10 px-4 py-10 sm:px-6 sm:py-12">
        <header class="flex flex-col gap-3 text-center lg:text-left">
            <p class="text-sm tracking-[0.35em] text-blue-200/80 uppercase">
                Blind Party Prototype
            </p>
            <h1
                class="font-minecraft text-3xl tracking-wider text-yellow-300 uppercase drop-shadow-[4px_4px_0px_rgba(0,0,0,0.65)] sm:text-4xl"
            >
                Game ID: <span class="text-white">{params.gameId}</span>
            </h1>
            <p class="text-base text-blue-100/80">
                Preview the arena layout and lobby roster while we hook up the live game feed.
            </p>
        </header>

        <CountdownBar duration={90} fillColor="#facc15" />

        <CountdownOverlay {remainingSeconds} />

        <div class="flex flex-col gap-8 lg:flex-row">
            <GameBoardPanel
                {mapSize}
                players={otherPlayersOnBoard}
                selfPlayer={selfPlayerOnBoard}
            />
            <!-- Show movement controls between board and roster on mobile -->
            <div class="lg:hidden">
                <PlayerMovementControls />
            </div>
            <PlayerRoster bind:players selfPlayer={selfPlayerSummary} />
        </div>

        <!-- Keep the original controls visible only on large screens (desktop) -->
        <div class="hidden lg:block">
            <PlayerMovementControls />
        </div>
    </div>
</div>
