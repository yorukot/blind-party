<script lang="ts">
	import GameBoardPanel from '$lib/components/game/game-board-panel.svelte';
	import PlayerMovementControls, {
		type Direction
	} from '$lib/components/game/player-movement-controls.svelte';
	import PlayerRoster, { type PlayerSummary } from '$lib/components/game/player-roster.svelte';

	interface Props {
		params: {
			gameId: string;
		};
	}

	let { params }: Props = $props();
	let mapSize = $state(18);
	let players = $state<PlayerSummary[]>([
		{
			id: '1',
			name: 'PixelPanda',
			status: 'ingame',
			accent: 'from-emerald-400 to-emerald-600'
		},
		{ id: '2', name: 'ShadowFox', status: 'ingame', accent: 'from-blue-400 to-indigo-600' },
		{ id: '3', name: 'NeonNova', status: 'spectating', accent: 'from-pink-400 to-rose-600' },
		{ id: '4', name: 'ByteKnight', status: 'eliminated', accent: 'from-slate-500 to-slate-700' }
	]);

	let lastMove = $state<Direction | null>(null);

	function handlePlayerMove(direction: Direction) {
		lastMove = direction;
	}

	$inspect(players);
</script>

<div class="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 text-white">
	<div class="mx-auto flex max-w-6xl flex-col gap-10 px-4 py-10 sm:px-6 sm:py-12">
		<header class="flex flex-col gap-3 text-center lg:text-left">
			<p class="text-sm tracking-[0.35em] text-blue-200/80 uppercase">
				Blind Party Prototype
			</p>
			<h1
				class="font-minecraft text-3xl tracking-wider text-yellow-300 uppercase drop-shadow-[4px_4px_0px_rgba(0,0,0,0.65)] sm:text-4xl"
			>
				Game ID: <span class="text-white">{params.gameId}</span>
			</h1>
			<p class="text-base text-blue-100/80">
				Preview the arena layout and lobby roster while we hook up the live game feed.
			</p>
		</header>

		<div class="flex flex-col gap-8 lg:flex-row">
			<GameBoardPanel {mapSize} />
			<PlayerRoster {players} />
		</div>
	</div>
</div>
