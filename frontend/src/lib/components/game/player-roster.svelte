<script lang="ts">
	export interface PlayerSummary {
		id: string;
		name: string;
		status: 'ready' | 'ingame' | 'eliminated';
		accent: string;
	}

	interface Props {
		players: PlayerSummary[];
	}

	let { players }: Props = $props();

	let totalPlayers = $derived(players.length);
	let activePlayers = $derived(players.filter((player) => player.status !== 'eliminated').length);
</script>

<aside class="w-full max-w-full space-y-6 rounded-3xl border-4 border-black bg-slate-950/85 p-6 shadow-[0_12px_0px_rgba(0,0,0,0.55)] backdrop-blur lg:max-w-sm">
	<div class="space-y-1">
		<h2 class="font-minecraft text-2xl uppercase tracking-wide text-cyan-300">Players</h2>
		<p class="text-sm text-blue-100/70">
			Players sync in real-time once the lobby is live. Showing sample roster for styling.
		</p>
	</div>

	<div class="flex items-center justify-between rounded-lg border-2 border-black bg-gradient-to-r from-emerald-500/20 via-transparent to-cyan-500/20 px-4 py-3 text-xs uppercase tracking-[0.25em] text-blue-100/60">
		<span>{activePlayers} active</span>
		<span>{totalPlayers} total</span>
	</div>

	<ul class="space-y-3">
		{#each players as player (player.id)}
			<li class="pixel-card flex items-center justify-between gap-3 rounded-xl border-2 border-black bg-slate-900/80 px-4 py-3 shadow-[4px_4px_0px_rgba(0,0,0,0.6)]">
				<div class="flex items-center gap-3">
					<span
						class={`h-10 w-10 rounded-lg border-2 border-black bg-gradient-to-br ${player.accent} shadow-[2px_2px_0px_rgba(0,0,0,0.5)]`}
					></span>
					<div>
						<p class="font-minecraft text-lg uppercase tracking-wide text-yellow-100">
							{player.name}
						</p>
						<p class={`text-xs uppercase tracking-[0.25em] ${player.status === 'eliminated' ? 'text-rose-300/70' : player.status === 'ingame' ? 'text-emerald-200/80' : 'text-blue-100/70'}`}>
							{player.status === 'ingame' ? 'On Mission' : player.status === 'ready' ? 'Ready' : 'Eliminated'}
						</p>
					</div>
				</div>
				<button
					class="rounded-md border-2 border-black bg-slate-800/80 px-3 py-1 text-[0.65rem] uppercase tracking-[0.3em] text-blue-100/70 shadow-[2px_2px_0px_rgba(0,0,0,0.45)] transition-colors hover:bg-slate-700"
					type="button"
				>
					Inspect
				</button>
			</li>
		{/each}
	</ul>
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
