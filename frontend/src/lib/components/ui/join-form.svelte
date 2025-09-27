<script lang="ts">
    import PixelButton from '$lib/components/ui/pixel-button.svelte';
    import PixelInput from '$lib/components/ui/pixel-input.svelte';

    interface Props {
        params: {
            gameId: string;
        };
        username: string;
        isConnecting: boolean;
        connectionError: string | null;
        connectionState: string;
        onJoin: () => void;
    }

    let {
        params,
        username = $bindable(),
        isConnecting,
        connectionError,
        connectionState,
        onJoin
    }: Props = $props();
</script>

<div class="flex min-h-screen flex-col items-center justify-center p-8">
    <div class="flex w-full max-w-md flex-col items-center space-y-8">
        <header class="text-center">
            <p class="mb-2 text-sm tracking-[0.35em] text-blue-200/80 uppercase">Blind Party</p>
            <h1
                class="font-minecraft text-3xl tracking-wider text-yellow-300 uppercase drop-shadow-[4px_4px_0px_rgba(0,0,0,0.65)]"
            >
                Game: <span class="text-white">{params.gameId}</span>
            </h1>
        </header>

        <div class="w-full space-y-4">
            <PixelInput
                bind:value={username}
                placeholder="Enter your username"
                maxlength={20}
            />

            {#if connectionError}
                <p class="text-sm text-red-400">{connectionError}</p>
            {/if}

            <PixelButton disabled={!username.trim() || isConnecting} onclick={onJoin}>
                {isConnecting ? 'Joining...' : 'Join Game'}
            </PixelButton>

            <div class="text-center">
                <p class="text-sm text-slate-400">
                    Connection: <span class="text-blue-300">{connectionState}</span>
                </p>
            </div>
        </div>
    </div>
</div>
