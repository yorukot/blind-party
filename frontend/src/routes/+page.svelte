<script lang="ts">
	import { goto } from '$app/navigation';
	import logo from '$lib/assets/blind-party.png';
	import PixelButton from '../lib/components/ui/PixelButton.svelte';
	import PixelInput from '../lib/components/ui/PixelInput.svelte';

	let gameId = $state('');

	function joinGame() {
		if (gameId.trim()) {
			goto(`/game/${gameId.trim()}`);
		}
	}
</script>

<div
	class="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 p-8"
>
	<div class="flex w-full max-w-md flex-col items-center space-y-8">
		<img src={logo} alt="Blind Party Logo" class="pixelated h-auto w-[32rem]" />

		<div class="w-full space-y-4">
			<PixelInput bind:value={gameId} placeholder="Enter Game ID" maxlength={10} />

			<PixelButton variant="primary" disabled={!gameId.trim()} onclick={joinGame}>
				Join Game
			</PixelButton>
		</div>
	</div>
</div>

<style>
	.pixelated {
		image-rendering: pixelated;
		image-rendering: -moz-crisp-edges;
		image-rendering: crisp-edges;
	}
</style>
