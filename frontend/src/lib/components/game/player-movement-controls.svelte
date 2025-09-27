<script lang="ts">
    import { onMount } from 'svelte';
    import { playerState, type Direction } from '$lib/player-state.svelte.js';

    interface Props {
        disabled?: boolean;
    }

    let { disabled = false }: Props = $props();

    function handlePointerDown(direction: Direction, event: PointerEvent) {
        if (disabled) {
            return;
        }

        event.preventDefault();
        playerState.triggerMove(direction);
    }

    function handlePointerEnd(direction: Direction, event: PointerEvent) {
        if (disabled) {
            return;
        }

        event.preventDefault();
        playerState.clearDirection(direction);
    }

    onMount(() => {
        function handleKeyDown(event: KeyboardEvent) {
            if (disabled) {
                return;
            }
            playerState.handleKeyDown(event);
        }

        function handleKeyUp(event: KeyboardEvent) {
            playerState.handleKeyUp(event);
        }

        window.addEventListener('keydown', handleKeyDown);
        window.addEventListener('keyup', handleKeyUp);

        return () => {
            window.removeEventListener('keydown', handleKeyDown);
            window.removeEventListener('keyup', handleKeyUp);
            playerState.clearPressedKeys();
        };
    });
</script>

<section
    class="movement-panel w-full rounded-3xl border-4 border-black bg-slate-950/85 px-6 py-8 text-blue-100 shadow-[0_16px_0px_rgba(0,0,0,0.55)] backdrop-blur"
    aria-label="Player movement controls"
>
    <div class="flex flex-col items-center gap-6">
        <div
            class="flex w-full flex-col items-start justify-between gap-2 sm:flex-row sm:items-center"
        >
            <h2 class="font-minecraft text-xl tracking-[0.35em] text-cyan-200 uppercase">
                Movement
            </h2>
            <p class="text-xs tracking-[0.3em] text-blue-200/70 uppercase">
                Use Arrow Keys or WASD
            </p>
        </div>

        <div class="grid w-full max-w-sm grid-cols-3 grid-rows-3 gap-3">
            <div></div>
            <button
                type="button"
                class={`inline-flex h-16 w-16 items-center justify-center rounded-2xl border-3 border-black text-xs tracking-wider text-blue-100 uppercase transition-all duration-100 ease-in-out ${
                    playerState.activeDirections.has('up')
                        ? 'translate-y-0.5 bg-gradient-to-br from-red-400/60 to-red-600/80 text-orange-50 shadow-[2px_2px_0_rgba(0,0,0,0.6)]'
                        : 'bg-gradient-to-br from-blue-400/40 to-blue-600/75 shadow-[4px_4px_0_rgba(0,0,0,0.55)] hover:shadow-[3px_3px_0_rgba(0,0,0,0.6)]'
                } focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-white/85`}
                onpointerdown={(event) => handlePointerDown('up', event)}
                onpointerup={(event) => handlePointerEnd('up', event)}
                onpointerleave={(event) => handlePointerEnd('up', event)}
                onpointercancel={(event) => handlePointerEnd('up', event)}
                aria-label="Move up"
            >
                <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    aria-hidden="true"
                    class="h-5 w-5 drop-shadow-[2px_2px_0_rgba(0,0,0,0.4)]"
                >
                    <path d="M12 5.5 5 13h4v6h6v-6h4z" />
                </svg>
            </button>
            <div></div>

            <button
                type="button"
                class={`inline-flex h-16 w-16 items-center justify-center rounded-2xl border-3 border-black text-xs tracking-wider text-blue-100 uppercase transition-all duration-100 ease-in-out ${
                    playerState.activeDirections.has('left')
                        ? 'translate-y-0.5 bg-gradient-to-br from-red-400/60 to-red-600/80 text-orange-50 shadow-[2px_2px_0_rgba(0,0,0,0.6)]'
                        : 'bg-gradient-to-br from-blue-400/40 to-blue-600/75 shadow-[4px_4px_0_rgba(0,0,0,0.55)] hover:shadow-[3px_3px_0_rgba(0,0,0,0.6)]'
                } focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-white/85`}
                onpointerdown={(event) => handlePointerDown('left', event)}
                onpointerup={(event) => handlePointerEnd('left', event)}
                onpointerleave={(event) => handlePointerEnd('left', event)}
                onpointercancel={(event) => handlePointerEnd('left', event)}
                aria-label="Move left"
            >
                <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    aria-hidden="true"
                    class="h-5 w-5 drop-shadow-[2px_2px_0_rgba(0,0,0,0.4)]"
                >
                    <path d="M5.5 12 13 19v-4h6v-6h-6V5z" />
                </svg>
            </button>

            <button
                type="button"
                class="pointer-events-none inline-flex h-16 w-16 items-center justify-center rounded-2xl border-3 border-black bg-slate-900/80 text-blue-200/50 opacity-50 shadow-none"
                disabled
                aria-hidden="true"
            >
                <span class="text-[0.6rem] tracking-[0.3em] uppercase">Move</span>
            </button>

            <button
                type="button"
                class={`inline-flex h-16 w-16 items-center justify-center rounded-2xl border-3 border-black text-xs tracking-wider text-blue-100 uppercase transition-all duration-100 ease-in-out ${
                    playerState.activeDirections.has('right')
                        ? 'translate-y-0.5 bg-gradient-to-br from-red-400/60 to-red-600/80 text-orange-50 shadow-[2px_2px_0_rgba(0,0,0,0.6)]'
                        : 'bg-gradient-to-br from-blue-400/40 to-blue-600/75 shadow-[4px_4px_0_rgba(0,0,0,0.55)] hover:shadow-[3px_3px_0_rgba(0,0,0,0.6)]'
                } focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-white/85`}
                onpointerdown={(event) => handlePointerDown('right', event)}
                onpointerup={(event) => handlePointerEnd('right', event)}
                onpointerleave={(event) => handlePointerEnd('right', event)}
                onpointercancel={(event) => handlePointerEnd('right', event)}
                aria-label="Move right"
            >
                <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    aria-hidden="true"
                    class="h-5 w-5 drop-shadow-[2px_2px_0_rgba(0,0,0,0.4)]"
                >
                    <path d="M18.5 12 11 5v4H5v6h6v4z" />
                </svg>
            </button>

            <div></div>
            <button
                type="button"
                class={`inline-flex h-16 w-16 items-center justify-center rounded-2xl border-3 border-black text-xs tracking-wider text-blue-100 uppercase transition-all duration-100 ease-in-out ${
                    playerState.activeDirections.has('down')
                        ? 'translate-y-0.5 bg-gradient-to-br from-red-400/60 to-red-600/80 text-orange-50 shadow-[2px_2px_0_rgba(0,0,0,0.6)]'
                        : 'bg-gradient-to-br from-blue-400/40 to-blue-600/75 shadow-[4px_4px_0_rgba(0,0,0,0.55)] hover:shadow-[3px_3px_0_rgba(0,0,0,0.6)]'
                } focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-white/85`}
                onpointerdown={(event) => handlePointerDown('down', event)}
                onpointerup={(event) => handlePointerEnd('down', event)}
                onpointerleave={(event) => handlePointerEnd('down', event)}
                onpointercancel={(event) => handlePointerEnd('down', event)}
                aria-label="Move down"
            >
                <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    aria-hidden="true"
                    class="h-5 w-5 drop-shadow-[2px_2px_0_rgba(0,0,0,0.4)]"
                >
                    <path d="M12 18.5 19 11h-4V5H9v6H5z" />
                </svg>
            </button>
            <div></div>
        </div>

        <p class="text-xs text-blue-200/60">
            Tap or press directions to queue moves. Multiple keys can be pressed simultaneously.
            Keyboard input works even when the buttons are not focused.
        </p>
    </div>
</section>

<style>
    .movement-panel {
        text-rendering: optimizeSpeed;
        -webkit-font-smoothing: none;
        -moz-osx-font-smoothing: grayscale;
    }
</style>
