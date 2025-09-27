<script lang="ts">
    interface Props {
        duration?: number;
        label?: string;
        autoStart?: boolean;
        fillColor?: string;
    }

    let {
        duration = 90,
        label = 'Round Countdown',
        autoStart = true,
        fillColor = '#22c55e'
    }: Props = $props();

    let remaining = $state(Math.max(duration, 0));
    let progress = $derived(duration > 0 ? (remaining / duration) * 100 : 0);
    let formattedTime = $derived.by(() => {
        const totalSeconds = Math.max(remaining, 0);
        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    });
    let fillStyle = $derived.by(() => {
        return `width: ${progress}%; background-color: ${fillColor}; box-shadow: inset 0 -4px 0 rgba(0,0,0,0.35);`;
    });

    $effect(() => {
        if (!autoStart || duration <= 0) {
            return;
        }

        remaining = Math.min(remaining, duration);

        const interval = setInterval(() => {
            if (remaining === 0) {
                return;
            }

            remaining = Math.max(0, remaining - 1);
        }, 1000);

        return () => {
            clearInterval(interval);
        };
    });
</script>

<section
    class="relative w-full rounded-3xl border-4 border-yellow-400/70 bg-slate-950/80 p-4 shadow-[8px_8px_0_0_rgba(0,0,0,0.6)]"
>
    <div
        class="flex items-center justify-between text-xs tracking-[0.3em] text-yellow-200/90 uppercase"
    >
        <span
            class="font-minecraft text-sm text-yellow-300 drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]"
        >
            {label}
        </span>
        <span class="font-minecraft text-base text-white drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]">
            {formattedTime}
        </span>
    </div>
    <div
        class="relative mt-3 h-6 w-full overflow-hidden rounded-2xl border-2 border-black/70 bg-slate-900 shadow-[inset_0_0_0_2px_rgba(0,0,0,0.8)]"
    >
        <div class="h-full transition-[width] duration-1000 ease-linear" style={fillStyle}></div>
        <div
            class="pointer-events-none absolute inset-0"
            style="background-image: repeating-linear-gradient(0deg, rgba(255,255,255,0.18) 0px, rgba(255,255,255,0.18) 2px, transparent 2px, transparent 4px); mix-blend-mode: screen;"
        ></div>
    </div>
    <div class="pointer-events-none absolute inset-0 rounded-3xl border-2 border-black/60"></div>
</section>
