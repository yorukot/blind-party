<script lang="ts">
    import { getBlockName } from '$lib/types/game';

    interface Props {
        targetBlock: number;
    }

    let { targetBlock }: Props = $props();

    const textureModules = import.meta.glob('$lib/assets/blocks/*.png', {
        as: 'url',
        eager: true
    });

    const textureUrls = Object.entries(textureModules)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([, url]) => url as string);

    let blockName = $derived(getBlockName(targetBlock));
    let blockTexturePath = $derived.by(() => {
        const textureIndex = Math.min(targetBlock, textureUrls.length - 1);
        return textureUrls[textureIndex] || textureUrls[0];
    });
</script>

<div
    class="absolute top-4 right-4 z-10 flex flex-col items-center gap-2 rounded-xl border-4 border-black bg-slate-900/50 p-3 shadow-[4px_4px_0_rgba(0,0,0,0.4)] opacity-50"
    aria-label={`Target: ${blockName} Block`}
>
    <div class="text-center">
        <p class="font-minecraft text-xs uppercase tracking-wider text-yellow-300 drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]">
            Target
        </p>
    </div>
    <div class="relative">
        <img
            src={blockTexturePath}
            alt={`${blockName} block`}
            class="h-12 w-12 rounded border-2 border-black shadow-[2px_2px_0_rgba(0,0,0,0.6)]"
            style="image-rendering: pixelated; image-rendering: crisp-edges;"
        />
        <div class="absolute inset-0 rounded border border-white/20"></div>
    </div>
    <p class="font-minecraft text-xs text-white drop-shadow-[1px_1px_0_rgba(0,0,0,0.8)]">
        {blockName}
    </p>
</div>