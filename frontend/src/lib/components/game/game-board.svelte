<script lang="ts">
    import { onMount } from 'svelte';
    import type { PlayerOnBoard } from '$lib/types/player';
    import { BLOCK_TEXTURE_NAMES } from '$lib/constants/block-textures';

    /**
     * Canvas-based renderer for the Blind Party map.
     * Displays the game map data.
     */
    interface Props {
        mapSize: number;
        gameMap?: number[][];
        tileSize?: number;
        players?: PlayerOnBoard[];
        selfPlayer?: PlayerOnBoard | null;
    }

    let { mapSize, gameMap = [], tileSize = 32, players = [], selfPlayer = null }: Props = $props();

    // Use centralized block texture names to dynamically import textures
    const textureUrls = BLOCK_TEXTURE_NAMES.map(name => {
        const modules = import.meta.glob('$lib/assets/blocks/*.png', { as: 'url', eager: true });
        const path = `/src/lib/assets/blocks/${name}.png`;
        return modules[path] as string;
    }).filter(Boolean);

    let normalizedMapSize = $derived.by(() => Math.max(1, Math.floor(mapSize ?? 0)));
    let safeTileSize = $derived.by(() => Math.max(8, Math.floor(tileSize ?? 0) || 32));
    let cssSize = $derived(normalizedMapSize * safeTileSize);

    let canvas: HTMLCanvasElement | null = null;
    let devicePixelRatio = $state(1);
    let texturesReady = $state(false);
    let mapGrid = $state<number[][]>([]);
    let blockImages: HTMLImageElement[] = [];

    const clampToBoard = (value: number, size: number) => Math.min(Math.max(value, 0), size - 1);

    const mapPlayerToToken = (player: PlayerOnBoard, isSelf = false) => {
        const x = clampToBoard(player.position.x, normalizedMapSize);
        const y = clampToBoard(player.position.y, normalizedMapSize);

        return {
            id: player.id,
            name: player.name,
            status: player.status,
            accent: player.accent,
            x,
            y,
            left: (x + 0.5) * safeTileSize,
            top: (y + 0.5) * safeTileSize,
            isSelf
        };
    };

    let otherPlayerTokens = $derived(
        players
            .filter((player) => player.status !== 'eliminated')
            .map((player) => mapPlayerToToken(player))
    );

    let selfPlayerToken = $derived.by(() => {
        if (!selfPlayer || selfPlayer.status === 'eliminated') {
            return null;
        }
        return mapPlayerToToken(selfPlayer, true);
    });

    $effect(() => {
        if (!import.meta.env.DEV) {
            return;
        }

        console.debug('[GameBoard] tokens derived', {
            otherPlayers: otherPlayerTokens.map((token) => ({
                id: token.id,
                x: token.x,
                y: token.y
            })),
            self: selfPlayerToken
                ? {
                      id: selfPlayerToken.id,
                      x: selfPlayerToken.x,
                      y: selfPlayerToken.y
                  }
                : null
        });
    });

    let playerDiameter = $derived(
        Math.min(safeTileSize, Math.max(16, Math.round(safeTileSize * 0.85)))
    );

    const getPlayerClasses = (status: PlayerOnBoard['status'], isSelf: boolean) => {
        if (isSelf) {
            return 'ring-4 ring-amber-300/90 ring-offset-2 ring-offset-black';
        }

        if (status === 'spectating') {
            return 'opacity-80';
        }

        return '';
    };

    $effect(() => {
        // Use provided map data
        if (gameMap && gameMap.length > 0) {
            mapGrid = gameMap;
        } else {
            // Clear map if no data is available
            mapGrid = [];
        }
    });

    function createImage(url: string) {
        return new Promise<HTMLImageElement>((resolve, reject) => {
            const img = new Image();
            img.onload = () => resolve(img);
            img.onerror = (event) => reject(event);
            img.src = url;
        });
    }

    function draw(grid: number[][], tile: number, ratio: number) {
        if (!canvas || !texturesReady || !grid.length) {
            return;
        }

        const context = canvas.getContext('2d');
        if (!context) {
            return;
        }

        // Reset the transform before drawing to keep scaling consistent.
        context.setTransform(1, 0, 0, 1, 0, 0);
        context.clearRect(0, 0, canvas.width, canvas.height);
        context.scale(ratio, ratio);

        for (let y = 0; y < grid.length; y += 1) {
            for (let x = 0; x < grid[y].length; x += 1) {
                const textureIndex = grid[y][x] % blockImages.length;
                const texture = blockImages[textureIndex];

                if (texture) {
                    context.drawImage(texture, x * tile, y * tile, tile, tile);
                } else {
                    context.fillStyle = '#000';
                    context.fillRect(x * tile, y * tile, tile, tile);
                }
            }
        }
    }

    onMount(() => {
        let cancelled = false;

        function updateDevicePixelRatio() {
            devicePixelRatio = window.devicePixelRatio || 1;
        }

        updateDevicePixelRatio();

        const handleResize = () => {
            updateDevicePixelRatio();
            draw(mapGrid, safeTileSize, devicePixelRatio);
        };

        window.addEventListener('resize', handleResize);

        Promise.all(textureUrls.map((url) => createImage(url)))
            .then((images) => {
                if (cancelled) {
                    return;
                }
                blockImages = images;
                texturesReady = true;
                draw(mapGrid, safeTileSize, devicePixelRatio);
            })
            .catch((error) => {
                console.error('Failed to load block textures', error);
            });

        return () => {
            cancelled = true;
            window.removeEventListener('resize', handleResize);
        };
    });

    let scaledCanvasSize = $derived(Math.round(cssSize * devicePixelRatio));

    $effect(() => {
        if (!texturesReady) {
            return;
        }

        draw(mapGrid, safeTileSize, devicePixelRatio);
    });
</script>

<div class="relative inline-block rounded-lg border-4 border-black bg-slate-900 p-2 shadow-xl">
    <div class="relative">
        <canvas
            bind:this={canvas}
            class="block-map h-auto w-full"
            width={scaledCanvasSize}
            height={scaledCanvasSize}
            style={`width: ${cssSize}px; height: ${cssSize}px;`}
        ></canvas>

        <div
            class="player-layer pointer-events-none absolute top-0 left-0"
            style={`width: ${cssSize}px; height: ${cssSize}px;`}
        >
            {#if selfPlayerToken}
                <div
                    class={`player-token absolute flex items-center justify-center rounded-full border-2 border-black bg-gradient-to-br ${selfPlayerToken.accent} text-white shadow-[3px_3px_0_rgba(0,0,0,0.6)] transition-transform duration-150 ease-out ${getPlayerClasses(
                        selfPlayerToken.status,
                        true
                    )}`}
                    style={`width: ${playerDiameter}px; height: ${playerDiameter}px; left: ${selfPlayerToken.left}px; top: ${selfPlayerToken.top}px; transform: translate(-50%, -50%);`}
                    aria-label={`${selfPlayerToken.name} (You)`}
                >
                    <span
                        class="font-minecraft text-xs tracking-widest drop-shadow-[1px_1px_0_rgba(0,0,0,0.65)]"
                    >
                        {selfPlayerToken.name.slice(0, 1).toUpperCase()}
                    </span>
                </div>
            {/if}

            {#each otherPlayerTokens as player (player.id)}
                <div
                    class={`player-token absolute flex items-center justify-center rounded-full border-2 border-black bg-gradient-to-br ${player.accent} text-white shadow-[3px_3px_0_rgba(0,0,0,0.6)] transition-transform duration-150 ease-out ${getPlayerClasses(
                        player.status,
                        false
                    )}`}
                    style={`width: ${playerDiameter}px; height: ${playerDiameter}px; left: ${player.left}px; top: ${player.top}px; transform: translate(-50%, -50%);`}
                    aria-label={`${player.name} (${player.status})`}
                >
                    <span
                        class="font-minecraft text-xs tracking-widest drop-shadow-[1px_1px_0_rgba(0,0,0,0.65)]"
                    >
                        {player.name.slice(0, 1).toUpperCase()}
                    </span>
                </div>
            {/each}
        </div>
    </div>
</div>

<style>
    .block-map {
        image-rendering: pixelated;
        image-rendering: crisp-edges;
        display: block;
    }

    .player-token {
        text-rendering: optimizeSpeed;
        -webkit-font-smoothing: none;
        -moz-osx-font-smoothing: grayscale;
    }
</style>
