<script lang="ts">
    import { onMount } from 'svelte';
    import GameBoard from './game-board.svelte';
    import type { PlayerOnBoard } from '$lib/types/player';

    interface Props {
        mapSize: number;
        targetSize?: number;
        minTileSize?: number;
        maxTileSize?: number;
        players?: PlayerOnBoard[];
        selfPlayer?: PlayerOnBoard | null;
    }

    let {
        mapSize,
        targetSize = 560,
        minTileSize = 10,
        maxTileSize = 48,
        players = [],
        selfPlayer = null
    }: Props = $props();

    let container: HTMLElement | null = null;
    let containerWidth = $state(targetSize);

    let safeMapSize = $derived.by(() => Math.max(1, Math.floor(mapSize ?? 1)));
    let effectiveTargetSize = $derived.by(() =>
        Math.max(minTileSize * safeMapSize, Math.min(containerWidth, targetSize))
    );
    let computedTileSize = $derived.by(() => {
        const tile = Math.floor(effectiveTargetSize / safeMapSize);
        return Math.max(minTileSize, Math.min(maxTileSize, tile));
    });

    onMount(() => {
        if (!container) {
            return;
        }

        const resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
                if (entry.target !== container) {
                    continue;
                }
                const width = entry.contentBoxSize?.[0]?.inlineSize ?? entry.contentRect.width;
                containerWidth = Math.max(width, minTileSize * safeMapSize);
            }
        });

        resizeObserver.observe(container);
        containerWidth = container.clientWidth;

        return () => {
            resizeObserver.disconnect();
        };
    });
</script>

<section
    bind:this={container}
    class="flex w-full items-center justify-center rounded-3xl border-4 border-black bg-slate-950/90 p-4 shadow-[0_16px_0px_rgba(0,0,0,0.55)] backdrop-blur"
>
    <GameBoard mapSize={safeMapSize} tileSize={computedTileSize} {players} {selfPlayer} />
</section>
