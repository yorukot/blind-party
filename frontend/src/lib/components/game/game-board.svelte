<script lang="ts">
    import { onMount } from 'svelte';

    /**
     * Canvas-based renderer for the Blind Party map.
     * Generates a random n x n block layout using the provided block textures.
     */
    interface Props {
        mapSize: number;
        tileSize?: number;
    }

    let { mapSize, tileSize = 32 }: Props = $props();

    const textureModules = import.meta.glob('$lib/assets/blocks/*.png', {
        as: 'url',
        eager: true
    });

    const textureUrls = Object.entries(textureModules)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([, url]) => url as string);

    let normalizedMapSize = $derived.by(() => Math.max(1, Math.floor(mapSize ?? 0)));
    let safeTileSize = $derived.by(() => Math.max(8, Math.floor(tileSize ?? 0) || 32));
    let cssSize = $derived(normalizedMapSize * safeTileSize);

    let canvas: HTMLCanvasElement | null = null;
    let devicePixelRatio = $state(1);
    let texturesReady = $state(false);
    let mapGrid = $state<number[][]>([]);
    let blockImages: HTMLImageElement[] = [];

    function generateMap(size: number) {
        const textureCount = textureUrls.length || 1;
        return Array.from({ length: size }, () =>
            Array.from({ length: size }, () => Math.floor(Math.random() * textureCount))
        );
    }

    $effect(() => {
        mapGrid = generateMap(normalizedMapSize);
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
    <canvas
        bind:this={canvas}
        class="block-map h-auto w-full"
        width={scaledCanvasSize}
        height={scaledCanvasSize}
        style={`width: ${cssSize}px; height: ${cssSize}px;`}
    ></canvas>
</div>

<style>
    .block-map {
        image-rendering: pixelated;
        image-rendering: crisp-edges;
        display: block;
    }
</style>
