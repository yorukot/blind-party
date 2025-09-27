<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary';
		disabled?: boolean;
		onclick?: () => void;
		children: import('svelte').Snippet;
	}

	let { variant = 'primary', disabled = false, onclick, children }: Props = $props();
</script>

<button
	class="
		pixel-button
		w-full
		{variant === 'primary' ? 'bg-gradient-to-b from-green-400 to-green-600 hover:from-green-300 hover:to-green-500' : 'bg-gradient-to-b from-blue-400 to-blue-600 hover:from-blue-300 hover:to-blue-500'}
		text-white
		border-4
		border-black
		px-6
		py-3
		text-xl
		font-minecraft
		font-bold
		uppercase
		tracking-wider
		disabled:opacity-50
		disabled:cursor-not-allowed
		active:scale-95
		transition-all
		duration-150
		pixel-text
	"
	style="
		box-shadow:
			inset -4px -4px 0px rgba(0,0,0,0.3),
			inset 4px 4px 0px rgba(255,255,255,0.3),
			4px 4px 8px rgba(0,0,0,0.4);
		text-shadow: 2px 2px 0px rgba(0,0,0,0.5);
	"
	{disabled}
	onclick={onclick}
>
	{@render children()}
</button>

<style>
	.pixel-button {
		text-rendering: optimizeSpeed;
		-webkit-font-smoothing: none;
		-moz-osx-font-smoothing: grayscale;
	}

	.pixel-button:hover:not(:disabled) {
		animation: pixel-bounce 0.6s ease-in-out infinite alternate;
	}

	.pixel-button:active:not(:disabled) {
		box-shadow:
			inset -2px -2px 0px rgba(0,0,0,0.3),
			inset 2px 2px 0px rgba(255,255,255,0.3),
			2px 2px 4px rgba(0,0,0,0.4);
	}

	@keyframes pixel-bounce {
		0% {
			transform: translateY(0px);
		}
		100% {
			transform: translateY(-2px);
		}
	}
</style>