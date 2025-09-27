# Blind Party Game - Agent Instructions

This is a web-based implementation of Hypixel's "Blind Party" minigame built with **Svelte 5**. This document provides comprehensive instructions for AI agents working on this codebase.

## Project Overview

Blind Party is a multiplayer party game where players navigate through various mini-games while "blind" (with limited vision or information). The game features real-time multiplayer functionality, dynamic game modes, and an engaging user interface.

## Technology Stack

- **Frontend Framework**: Svelte 5 (using runes syntax)
- **Language**: TypeScript
- **Styling**: TailwindCSS
- **Real-time Communication**: WebSockets
- **Build Tool**: Vite
- **Package Manager**: pnpm

## Svelte 5 Syntax Guidelines

### Component Structure

```svelte
<script lang="ts">
    // Use $state for reactive variables
    let count = $state(0);

    // Use $props for component properties
    let { title, required, optional = 'default' } = $props();

    // Use $derived for computed values (expressions only)
    // For functions, use $derived.by(() => {})
    let doubled = $derived(count * 2);

    // Use $effect for side effects
    $effect(() => {
        console.log(`Count is now: ${count}`);
    });
</script>

<div class="component">
    <h1>{title}</h1>
    <p>Count: {count}</p>
    <button onclick={() => count++}>Increment</button>
</div>
```

### Key Svelte 5 Changes

- Replace `export let` with `$props()` destructuring
- Replace `$:` reactive statements with `$derived` or `$effect`
- Use `$state` instead of regular variables for reactivity
- Use `$bindable()` for two-way binding props
- Components are dynamic by default (no need for `<svelte:component>`)

## Development Guidelines

### Component Creation

1. Always use Svelte 5 runes syntax
2. Follow TypeScript best practices
3. Use TailwindCSS for styling
4. Implement proper prop validation
5. Add JSDoc comments for complex components
6. Always check for existing components to reuse before creating new ones
7. Design many custom styled components for consistent UI

### State Management

- Use `$state` for local component state
- Create global state using classes with `$state` in `.svelte.js`/`.svelte.ts` files
- Use context API for deeply nested prop passing
- Maintain immutable state patterns

### Real-time Features

- Implement WebSocket connections in `src/lib/websocket.ts`
- Handle connection states gracefully
- Implement reconnection logic
- Sync game state across all clients

## Coding Standards

### Svelte Components

```svelte
<script lang="ts">
    // Props with proper typing
    interface Props {
        players: Player[];
        gameState: GameState;
        onPlayerAction?: (action: PlayerAction) => void;
    }

    let { players, gameState, onPlayerAction }: Props = $props();

    // Local reactive state
    let selectedPlayer = $state<Player | null>(null);

    // Derived values
    let alivePlayers = $derived(players.filter((p) => p.isAlive));

    // Effects for side effects
    $effect(() => {
        if (gameState === 'ended') {
            // Handle game end
        }
    });
</script>

<div class="game-component">
    {#each alivePlayers as player (player.id)}
        <PlayerCard
            {player}
            selected={selectedPlayer?.id === player.id}
            onclick={() => (selectedPlayer = player)}
        />
    {/each}
</div>
```

### TypeScript Usage

- Define interfaces for all data structures
- Use strict type checking
- Implement proper error handling
- Use generic types where appropriate

### Styling Guidelines

- Use TailwindCSS utility classes
- Implement responsive design patterns
- Follow a consistent color scheme
- Use CSS custom properties for theme variables

## Common Patterns

### Loading States

```svelte
<script lang="ts">
    let isLoading = $state(false);
    let data = $state(null);

    async function fetchData() {
        isLoading = true;
        try {
            data = await api.fetchGameData();
        } finally {
            isLoading = false;
        }
    }
</script>

{#if isLoading}
    <LoadingSpinner />
{:else if data}
    <GameContent {data} />
{:else}
    <ErrorMessage />
{/if}
```

### Event Handling

```svelte
<script lang="ts">
    let { onGameEvent } = $props();

    function handlePlayerMove(direction: Direction) {
        onGameEvent?.({
            type: 'player_move',
            direction,
            timestamp: Date.now()
        });
    }
</script>

<div class="controls">
    <button onclick={() => handlePlayerMove('up')}>�</button>
    <button onclick={() => handlePlayerMove('down')}>�</button>
</div>
```

## Testing Guidelines

### Component Testing

- Test component rendering with different props
- Test user interactions and state changes
- Mock external dependencies
- Test accessibility features

### Integration Testing

- Test game flow end-to-end
- Test WebSocket communication
- Test multiplayer synchronization
- Test error scenarios

## Performance Considerations

- Use `$derived` instead of recalculating values
- Implement proper key attributes in `{#each}` blocks
- Optimize WebSocket message handling
- Use lazy loading for heavy components
- Implement proper cleanup in `$effect`

## Security Guidelines

- Validate all user inputs
- Sanitize player names and messages
- Implement rate limiting
- Use secure WebSocket connections
- Never trust client-side game state

## Deployment

- Build with `pnpm run build` (only when explicitly requested)
- Test with `pnpm run preview` (only when explicitly requested)
- Ensure environment variables are configured
- Monitor performance and error rates
- Implement proper logging

**Important**: Never run build, dev, or preview commands unless explicitly requested by the user.

## Code Quality

- **CRITICAL**: Always run `pnpm check` after completing any task and before claiming it's done
- Solve ALL linter errors and type errors before marking tasks as complete
- Never submit code that fails linting or type checking
- If `pnpm check` is not available, ask the user for the correct lint/typecheck command

---

Remember: Always prioritize player experience, maintain clean code architecture, and follow Svelte 5 best practices throughout development.
