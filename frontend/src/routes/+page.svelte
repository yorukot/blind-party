<script lang="ts">
    import { goto } from '$app/navigation';
    import logo from '$lib/assets/blind-party.png';
    import PixelButton from '../lib/components/ui/pixel-button.svelte';
    import PixelInput from '../lib/components/ui/pixel-input.svelte';
    import { createGame as apiCreateGame, HttpApiError } from '$lib/api/http';

    let gameId = $state('');
    let isCreating = $state(false);
    let createError = $state<string | null>(null);

    function joinGame() {
        if (gameId.trim()) {
            goto(`/game/${gameId.trim()}`);
        }
    }

    async function createGame() {
        isCreating = true;
        createError = null;

        try {
            const response = await apiCreateGame();
            goto(`/game/${response.game_id}`);
        } catch (error) {
            if (error instanceof HttpApiError) {
                createError = error.message;
            } else {
                createError = error instanceof Error ? error.message : 'Failed to create game';
            }
        } finally {
            isCreating = false;
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

            <div class="flex items-center justify-center">
                <div class="h-px flex-1 bg-slate-600"></div>
                <span class="px-4 text-sm text-slate-400">OR</span>
                <div class="h-px flex-1 bg-slate-600"></div>
            </div>

            <PixelButton variant="secondary" disabled={isCreating} onclick={createGame}>
                {isCreating ? 'Creating...' : 'Create New Game'}
            </PixelButton>

            {#if createError}
                <p class="text-center text-sm text-red-400">{createError}</p>
            {/if}
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
