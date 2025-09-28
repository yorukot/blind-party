<script lang="ts">
    import { BLOCK_TEXTURE_NAMES } from '$lib/constants/block-textures';

    interface Props {
        targetBlock: number;
    }

    let { targetBlock }: Props = $props();

    // Use centralized block texture names to get the correct texture
    const textureUrls = BLOCK_TEXTURE_NAMES.map((name) => {
        const modules = import.meta.glob('$lib/assets/blocks/*.png', { as: 'url', eager: true });
        const path = `/src/lib/assets/blocks/${name}.png`;
        return modules[path] as string;
    }).filter(Boolean);

    let blockTexturePath = $derived.by(() => {
        const textureIndex = Math.min(targetBlock, textureUrls.length - 1);
        return textureUrls[textureIndex] || textureUrls[0];
    });
</script>

<div
    class="absolute top-4 right-4 z-10 flex flex-col items-center gap-2 rounded-xl border-4 border-black bg-slate-900/50 p-3 opacity-90 shadow-[4px_4px_0_rgba(0,0,0,0.4)]"
>
    <div class="text-center">
        <p
            class="font-minecraft text-xs tracking-wider text-yellow-300 uppercase drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]"
        >
            Target
        </p>
    </div>
    <div class="relative">
        <img
            src={blockTexturePath}
            alt="Target block texture"
            class="h-12 w-12 rounded border-2 border-black shadow-[2px_2px_0_rgba(0,0,0,0.6)]"
            style="image-rendering: pixelated; image-rendering: crisp-edges;"
        />
        <div class="absolute inset-0 rounded border border-white/20"></div>
    </div>
</div>
