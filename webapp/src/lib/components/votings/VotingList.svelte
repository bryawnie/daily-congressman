<script lang="ts">
	import type { Voting } from '#lib/types.ts';
	import VotingCard from './VotingCard.svelte';

	const INITIAL_COUNT = 2;
	const MAX_COUNT = 10;

	let { votings }: { votings: Voting[] } = $props();
	let expanded = $state(false);

	const latest = $derived(votings.slice(0, MAX_COUNT));
	const visible = $derived(expanded ? latest : latest.slice(0, INITIAL_COUNT));
</script>

{#if votings.length === 0}
	<p class="empty">No registra votaciones recientes.</p>
{:else}
	<ul>
		{#each visible as voting (voting.votingId)}
			<li><VotingCard {voting} /></li>
		{/each}
	</ul>
	{#if latest.length > INITIAL_COUNT}
		<button type="button" aria-expanded={expanded} onclick={() => (expanded = !expanded)}>
			{expanded ? 'Mostrar menos' : `Ver las últimas ${latest.length} votaciones`}
		</button>
	{/if}
{/if}

<style>
	ul {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 22rem), 1fr));
		gap: var(--space-3);
	}

	button {
		margin-block-start: var(--space-3);
		padding: 0;
		border: 0;
		background: none;
		color: var(--color-accent);
		font: inherit;
		font-size: 0.875rem;
		cursor: pointer;
	}

	button:hover {
		text-decoration: underline;
	}

	.empty {
		color: var(--color-text-muted);
	}
</style>
