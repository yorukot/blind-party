<script lang="ts">
    import { flip } from 'svelte/animate';
    import { slide } from 'svelte/transition';
    import type { PlayerSummary } from '$lib/types/player';

    interface Props {
        players: PlayerSummary[];
        selfPlayer?: PlayerSummary | null;
    }

    let { players = $bindable(), selfPlayer = null }: Props = $props();

    let selfPlayerId = $derived(selfPlayer?.id ?? null);

    const statusOrder = { ingame: 0, eliminated: 1, spectating: 2 } as const;

    let otherPlayersSorted = $derived.by(() =>
        players
            .filter((player) => player.id !== selfPlayerId)
            .toSorted((a, b) => statusOrder[a.status] - statusOrder[b.status])
    );

    let sortedPlayers = $derived.by(() =>
        selfPlayer ? [selfPlayer, ...otherPlayersSorted] : otherPlayersSorted
    );

    let totalPlayers = $derived(sortedPlayers.length);
    let activePlayers = $derived.by(
        () => sortedPlayers.filter((player) => player.status === 'ingame').length
    );
    let inactivePlayers = $derived.by(
        () => sortedPlayers.filter((player) => player.status !== 'ingame').length
    );
</script>

<aside
    class="w-full max-w-full space-y-6 rounded-3xl border-4 border-black bg-slate-950/85 p-6 shadow-[0_12px_0px_rgba(0,0,0,0.55)] backdrop-blur transition-all duration-300 lg:max-w-sm"
>
    <div class="space-y-1">
        <h2 class="font-minecraft text-2xl tracking-wide text-cyan-300 uppercase">Players</h2>
        <p class="text-sm text-blue-100/70">
            Players sync in real-time once the lobby is live. Showing sample roster for styling.
        </p>
    </div>

    <div
        class="flex items-center justify-between rounded-lg border-2 border-black bg-gradient-to-r from-emerald-500/20 via-transparent to-cyan-500/20 px-4 py-3 text-xs tracking-[0.25em] text-blue-100/60 uppercase"
    >
        <span>{activePlayers} active</span>
        <span>{totalPlayers} total</span>
    </div>

    <div class="space-y-3">
        {#each sortedPlayers as player, index (player.id)}
            <div animate:flip={{ duration: 300, easing: (t) => t * (2 - t) }}>
                <!-- Separator between active and inactive players -->
                {#if index > 0 && sortedPlayers[index - 1].status === 'ingame' && player.status !== 'ingame' && activePlayers > 0 && inactivePlayers > 0}
                    <div
                        class="mb-3 h-px bg-gradient-to-r from-transparent via-slate-600/50 to-transparent"
                        transition:slide={{ duration: 300, axis: 'y' }}
                    ></div>
                {/if}

                <div
                    class={`pixel-card flex items-center justify-between gap-3 rounded-xl border-2 border-black px-4 py-3 shadow-[4px_4px_0px_rgba(0,0,0,0.6)] transition-all duration-500 ease-in-out ${
                        player.id === selfPlayerId
                            ? 'bg-slate-900/90 ring-4 ring-amber-300/90 ring-offset-2 ring-offset-black'
                            : player.status === 'ingame'
                              ? 'bg-slate-900/80'
                              : 'bg-slate-900/50 opacity-75'
                    }`}
                >
                    <div class="flex items-center gap-3">
                        <span
                            class={`relative h-10 w-10 rounded-lg border-2 border-black bg-gradient-to-br ${player.accent} shadow-[2px_2px_0px_rgba(0,0,0,0.5)] transition-all duration-500 ease-in-out ${player.status !== 'ingame' ? 'grayscale' : ''}`}
                        ></span>
                        <div>
                            <p
                                class={`font-minecraft text-lg tracking-wide uppercase transition-all duration-500 ease-in-out ${
                                    player.status !== 'ingame'
                                        ? 'text-yellow-100/70'
                                        : 'text-yellow-100'
                                }`}
                            >
                                {player.name}
                            </p>
                            {#if player.id === selfPlayerId}
                                <p
                                    class="text-xs tracking-[0.25em] text-amber-200/80 uppercase transition-all duration-500 ease-in-out"
                                >
                                    You
                                </p>
                            {:else if player.status === 'eliminated'}
                                <p
                                    class="text-xs tracking-[0.25em] text-rose-300/70 uppercase transition-all duration-500 ease-in-out"
                                >
                                    Eliminated
                                </p>
                            {:else if player.status === 'ingame'}
                                <p
                                    class="text-xs tracking-[0.25em] text-emerald-200/80 uppercase transition-all duration-500 ease-in-out"
                                >
                                    On Mission
                                </p>
                            {:else if player.status === 'spectating'}
                                <p
                                    class="text-xs tracking-[0.25em] text-blue-100/70 uppercase transition-all duration-500 ease-in-out"
                                >
                                    Spectating
                                </p>
                            {/if}
                        </div>
                    </div>
                    <button
                        class={`cursor-pointer rounded-md border-2 border-black px-3 py-1 text-[0.65rem] tracking-[0.3em] uppercase shadow-[2px_2px_0px_rgba(0,0,0,0.45)] transition-all duration-500 ease-in-out ${
                            player.id === selfPlayerId
                                ? 'pointer-events-none bg-slate-800/70 text-amber-200/70 opacity-80'
                                : player.status === 'ingame'
                                  ? 'bg-slate-800/80 text-blue-100/70 hover:bg-slate-700'
                                  : 'bg-slate-800/50 text-blue-100/50'
                        }`}
                        type="button"
                        disabled={player.id === selfPlayerId || player.status !== 'ingame'}
                        onclick={() => {
                            if (player.id === selfPlayerId) {
                                return;
                            }

                            if (player.status === 'ingame') {
                                const originalIndex = players.findIndex((p) => p.id === player.id);
                                if (originalIndex !== -1) {
                                    players[originalIndex].status = 'eliminated';
                                }
                            }
                        }}
                    >
                        Kill
                    </button>
                </div>
            </div>
        {/each}
    </div>
</aside>

<style>
    .pixel-card {
        text-rendering: optimizeSpeed;
        -webkit-font-smoothing: none;
        -moz-osx-font-smoothing: grayscale;
        transition: transform 0.15s ease;
    }

    .pixel-card:hover {
        transform: translateY(-2px);
    }
</style>
