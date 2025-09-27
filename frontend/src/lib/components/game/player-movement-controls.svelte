<script lang="ts">
	import { onMount } from 'svelte';

	export type Direction = 'up' | 'down' | 'left' | 'right';

	interface Props {
		disabled?: boolean;
		onMove?: (direction: Direction) => void;
	}

	let { disabled = false, onMove }: Props = $props();

	let activeDirection = $state<Direction | null>(null);

	const keyDirectionMap: Record<string, Direction> = {
		ArrowUp: 'up',
		ArrowDown: 'down',
		ArrowLeft: 'left',
		ArrowRight: 'right',
		w: 'up',
		s: 'down',
		a: 'left',
		d: 'right'
	};

	const pressedKeys = new Set<string>();

	function getDirectionFromKey(key: string): Direction | undefined {
		if (!key) {
			return undefined;
		}

		if (keyDirectionMap[key]) {
			return keyDirectionMap[key];
		}

		const normalized = key.toLowerCase();
		return keyDirectionMap[normalized];
	}

	function triggerMove(direction: Direction) {
		if (disabled) {
			return;
		}

		activeDirection = direction;
		onMove?.(direction);
	}

	function clearDirection(direction: Direction | null) {
		if (direction && activeDirection === direction) {
			activeDirection = null;
		}
	}

	function handlePointerDown(direction: Direction, event: PointerEvent) {
		if (disabled) {
			return;
		}

		event.preventDefault();
		triggerMove(direction);
	}

	function handlePointerEnd(direction: Direction, event: PointerEvent) {
		if (disabled) {
			return;
		}

		event.preventDefault();
		clearDirection(direction);
	}

	onMount(() => {
		function handleKeyDown(event: KeyboardEvent) {
			if (disabled) {
				return;
			}

			const direction = getDirectionFromKey(event.key);
			if (!direction) {
				return;
			}

			if (pressedKeys.has(event.key)) {
				// Avoid repeatedly triggering while the key is held down.
				event.preventDefault();
				return;
			}

			pressedKeys.add(event.key);
			event.preventDefault();
			triggerMove(direction);
		}

		function handleKeyUp(event: KeyboardEvent) {
			const direction = getDirectionFromKey(event.key);
			pressedKeys.delete(event.key);
			if (!direction) {
				return;
			}

			clearDirection(direction);
		}

		window.addEventListener('keydown', handleKeyDown);
		window.addEventListener('keyup', handleKeyUp);

		return () => {
			window.removeEventListener('keydown', handleKeyDown);
			window.removeEventListener('keyup', handleKeyUp);
			pressedKeys.clear();
		};
	});
</script>

<section
	class="movement-panel w-full rounded-3xl border-4 border-black bg-slate-950/85 px-6 py-8 text-blue-100 shadow-[0_16px_0px_rgba(0,0,0,0.55)] backdrop-blur"
	aria-label="Player movement controls"
>
	<div class="flex flex-col items-center gap-6">
		<div class="flex w-full flex-col items-start justify-between gap-2 sm:flex-row sm:items-center">
			<h2 class="font-minecraft text-xl uppercase tracking-[0.35em] text-cyan-200">Movement</h2>
			<p class="text-xs uppercase tracking-[0.3em] text-blue-200/70">
				Use Arrow Keys or WASD
			</p>
		</div>

		<div class="grid w-full max-w-sm grid-cols-3 grid-rows-3 gap-3">
			<div></div>
			<button
				type="button"
				class={`control-button ${activeDirection === 'up' ? 'is-active' : ''}`}
				onpointerdown={(event) => handlePointerDown('up', event)}
				onpointerup={(event) => handlePointerEnd('up', event)}
				onpointerleave={(event) => handlePointerEnd('up', event)}
				onpointercancel={(event) => handlePointerEnd('up', event)}
				aria-label="Move up"
			>
				<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="h-5 w-5">
					<path d="M12 5.5 5 13h4v6h6v-6h4z" />
				</svg>
			</button>
			<div></div>

			<button
				type="button"
				class={`control-button ${activeDirection === 'left' ? 'is-active' : ''}`}
				onpointerdown={(event) => handlePointerDown('left', event)}
				onpointerup={(event) => handlePointerEnd('left', event)}
				onpointerleave={(event) => handlePointerEnd('left', event)}
				onpointercancel={(event) => handlePointerEnd('left', event)}
				aria-label="Move left"
			>
				<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="h-5 w-5">
					<path d="M5.5 12 13 19v-4h6v-6h-6V5z" />
				</svg>
			</button>

			<button
				type="button"
				class="control-button center"
				disabled
				aria-hidden="true"
			>
				<span class="text-[0.6rem] uppercase tracking-[0.3em] text-blue-200/70">Move</span>
			</button>

			<button
				type="button"
				class={`control-button ${activeDirection === 'right' ? 'is-active' : ''}`}
				onpointerdown={(event) => handlePointerDown('right', event)}
				onpointerup={(event) => handlePointerEnd('right', event)}
				onpointerleave={(event) => handlePointerEnd('right', event)}
				onpointercancel={(event) => handlePointerEnd('right', event)}
				aria-label="Move right"
			>
				<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="h-5 w-5">
					<path d="M18.5 12 11 5v4H5v6h6v4z" />
				</svg>
			</button>

			<div></div>
			<button
				type="button"
				class={`control-button ${activeDirection === 'down' ? 'is-active' : ''}`}
				onpointerdown={(event) => handlePointerDown('down', event)}
				onpointerup={(event) => handlePointerEnd('down', event)}
				onpointerleave={(event) => handlePointerEnd('down', event)}
				onpointercancel={(event) => handlePointerEnd('down', event)}
				aria-label="Move down"
			>
				<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="h-5 w-5">
					<path d="M12 18.5 19 11h-4V5H9v6H5z" />
				</svg>
			</button>
			<div></div>
		</div>

		<p class="text-xs text-blue-200/60">
			Tap or press a direction to queue a move. Keyboard input works even when the buttons are not focused.
		</p>
	</div>
</section>

<style>
	.movement-panel {
		text-rendering: optimizeSpeed;
		-webkit-font-smoothing: none;
		-moz-osx-font-smoothing: grayscale;
	}

	.control-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		height: 4rem;
		width: 4rem;
		border-radius: 1rem;
		border-width: 3px;
		border-style: solid;
		border-color: #000;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.4), rgba(37, 99, 235, 0.75));
		box-shadow: 4px 4px 0 rgba(0, 0, 0, 0.55);
		transition: transform 0.1s ease, box-shadow 0.1s ease, background 0.2s ease;
		color: #e0f2fe;
		text-transform: uppercase;
		font-family: 'Inter', sans-serif;
		font-size: 0.75rem;
		letter-spacing: 0.2em;
	}

	.control-button svg {
		filter: drop-shadow(2px 2px 0 rgba(0, 0, 0, 0.4));
	}

	.control-button:active,
	.control-button.is-active {
		transform: translateY(2px);
		box-shadow: 2px 2px 0 rgba(0, 0, 0, 0.6);
		background: linear-gradient(135deg, rgba(248, 113, 113, 0.6), rgba(239, 68, 68, 0.8));
		color: #fff7ed;
	}

	.control-button:focus-visible {
		outline: 3px solid rgba(255, 255, 255, 0.85);
		outline-offset: 2px;
	}

	.control-button[disabled] {
		opacity: 0.5;
		box-shadow: none;
		pointer-events: none;
	}

	.control-button.center {
		background: rgba(15, 23, 42, 0.8);
		color: rgba(191, 219, 254, 0.5);
		box-shadow: inset 0 0 0px 1px rgba(148, 163, 184, 0.25);
	}
</style>
