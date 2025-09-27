<script lang="ts">
    interface Props {
        /** Current second(s) remaining in the final countdown window. */
        remainingSeconds: number;
        /** Allow the parent to explicitly toggle visibility. */
        visible?: boolean;
        /** Maximum countdown window the overlay should cover. */
        windowSize?: number;
        /** Optional headline text displayed above the timer. */
        label?: string;
        /** Supporting message displayed below the timer. */
        message?: string;
        /** Accent colour used for borders, glow and the radial progress ring. */
        fillColor?: string;
    }

    let {
        remainingSeconds,
        visible = true,
        windowSize = 5,
        label = 'Final Countdown',
        message = 'Round ending soon — get ready!',
        fillColor = '#fef08a'
    }: Props = $props();

    /** Clamp the timer so visuals stay within the configured window. */
    let clamped = $derived(Math.max(Math.min(remainingSeconds, windowSize), 0));
    let shouldRender = $derived(
        visible &&
            remainingSeconds > 0 &&
            remainingSeconds <= windowSize &&
            Number.isFinite(remainingSeconds)
    );
    let intRemaining = $derived(Math.ceil(clamped));
    let progress = $derived(windowSize > 0 ? clamped / windowSize : 0);
    let overlayStyle = $derived.by(() => {
        const safeProgress = Math.min(Math.max(progress, 0), 1);
        const intensity = 0.12 + (1 - safeProgress) * 0.18;
        const glow = `box-shadow: 0 22px 68px color-mix(in srgb, ${fillColor} ${(intensity * 100).toFixed(0)}%, transparent);`;
        const borderColor = `color-mix(in srgb, var(--overlay-accent) 65%, rgba(148, 163, 184, 0.34))`;
        const backgroundGradient =
            'linear-gradient(145deg, ' +
            `color-mix(in srgb, var(--overlay-accent) 28%, rgba(30, 27, 75, 0.45)) 0%, ` +
            `color-mix(in srgb, var(--overlay-accent) 14%, rgba(15, 23, 42, 0.35)) 45%, ` +
            `rgba(2, 6, 23, 0.50) 100%)`;
        return `--overlay-accent: ${fillColor}; ${glow} border-color: ${borderColor}; background: ${backgroundGradient};`;
    });
    let labelStyle = $derived(`color: color-mix(in srgb, var(--overlay-accent) 80%, white 20%);`);
    let numberStyle = $derived(
        'color: color-mix(in srgb, var(--overlay-accent) 92%, white 8%); ' +
            'text-shadow: 6px 6px 0 rgba(0,0,0,0.45);'
    );
    let messageStyle = $derived(
        'color: color-mix(in srgb, var(--overlay-accent) 75%, rgba(255,255,255,0.85));'
    );
</script>

{#if shouldRender}
    <div class="pointer-events-none fixed inset-0 z-50 flex items-center justify-center">
        <div
            class="relative flex flex-col items-center gap-6 rounded-[32px] border-4 px-12 py-10 text-center text-white shadow-[0_18px_60px_rgba(0,0,0,0.4)]"
            style={overlayStyle}
        >
            <div
                class="overlay-sheen pointer-events-none absolute -inset-8 rounded-[40px] border-2"
            ></div>

            <span
                class="font-minecraft text-[0.85rem] tracking-[0.4em] uppercase drop-shadow-[3px_3px_0_rgba(0,0,0,0.4)]"
                style={labelStyle}
            >
                {label}
            </span>

            <div class="relative flex h-40 w-40 items-center justify-center">
                {#key intRemaining}
                    <span class="countdown-number font-minecraft text-8xl" style={numberStyle}>
                        {intRemaining}
                    </span>
                {/key}
            </div>

            <p
                class="text-sm font-medium drop-shadow-[2px_2px_0_rgba(0,0,0,0.4)]"
                style={messageStyle}
            >
                {message}
            </p>
        </div>
    </div>
{/if}

<style>
    .overlay-sheen {
        animation: overlay-pulse 1.4s ease-in-out infinite;
        border-color: color-mix(in srgb, var(--overlay-accent) 38%, rgba(255, 255, 255, 0.18));
        background: radial-gradient(transparent 70%);
    }

    .countdown-number {
        animation: overlay-pop 1s cubic-bezier(0.21, 1.12, 0.59, 1) forwards;
    }

    @keyframes overlay-pulse {
        0%,
        100% {
            opacity: 0.6;
            transform: scale(0.97);
        }

        50% {
            opacity: 0.95;
            transform: scale(1.01);
        }
    }

    @keyframes overlay-breathe {
        0%,
        100% {
            transform: scale(0.98);
            box-shadow: 0 0 12px rgba(248, 250, 252, 0.2);
        }

        50% {
            transform: scale(1.02);
            box-shadow: 0 0 26px rgba(248, 250, 252, 0.32);
        }
    }

    @keyframes overlay-pop {
        0% {
            opacity: 0;
            transform: scale(0.7);
            filter: blur(6px);
        }

        35% {
            opacity: 1;
            transform: scale(1.1);
            filter: blur(0px);
        }

        55% {
            transform: scale(0.95);
        }

        100% {
            transform: scale(1);
        }
    }
</style>
