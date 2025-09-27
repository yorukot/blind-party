<script lang="ts">
    interface Props {
        duration?: number;
        label?: string;
        autoStart?: boolean;
        fillColor?: string;
        borderColor?: string;
    }

    let {
        duration = 90,
        label = 'Round Countdown',
        autoStart = true,
        fillColor = '#22c55e',
        borderColor
    }: Props = $props();

    let remaining = $state(Math.max(duration, 0));
    let progress = $derived(duration > 0 ? (remaining / duration) * 100 : 0);
    let formattedTime = $derived.by(() => {
        const totalSeconds = Math.max(remaining, 0);
        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    });
    let resolvedBorderColor = $derived(borderColor ?? fillColor);
    let containerStyle = $derived.by(() => {
        const frameColor = `color-mix(in srgb, ${resolvedBorderColor} 55%, rgba(15, 23, 42, 0.52))`;
        const backgroundGradient =
            'linear-gradient(135deg, rgba(2, 6, 23, 0.95) 0%, ' +
            `color-mix(in srgb, ${fillColor} 12%, rgba(15, 23, 42, 0.82)) 38%, ` +
            `rgba(2, 6, 23, 0.92) 100%)`;
        return `--countdown-bar-fill: ${fillColor}; border-color: ${frameColor}; background: ${backgroundGradient};`;
    });
    let fillStyle = $derived.by(() => {
        const barGradient =
            'linear-gradient(270deg, ' +
            `color-mix(in srgb, var(--countdown-bar-fill) 96%, white 4%) 0%, ` +
            `color-mix(in srgb, var(--countdown-bar-fill) 78%, rgba(0, 0, 0, 0.08)) 100%)`;
        return `width: ${progress}%; background: ${barGradient}; box-shadow: inset 0 -4px 0 rgba(0,0,0,0.35);`;
    });
    let labelStyle = $derived(
        `color: color-mix(in srgb, var(--countdown-bar-fill) 82%, white 18%);`
    );
    let timeStyle = $derived(
        'color: color-mix(in srgb, var(--countdown-bar-fill) 94%, white 6%); ' +
            'text-shadow: 2px 2px 0 rgba(0,0,0,0.7);'
    );

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
    class="relative w-full rounded-3xl border-4 p-4 shadow-[8px_8px_0_0_rgba(0,0,0,0.6)]"
    style={containerStyle}
>
    <div class="flex items-center justify-between text-xs tracking-[0.3em] uppercase">
        <span
            class="font-minecraft text-sm drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]"
            style={labelStyle}
        >
            {label}
        </span>
        <span
            class="font-minecraft text-base drop-shadow-[2px_2px_0_rgba(0,0,0,0.7)]"
            style={timeStyle}
        >
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
