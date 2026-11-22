<script lang="ts">
	import type { Voting } from '#lib/types.ts';
	import VotingCard from './VotingCard.svelte';

	let { votings }: { votings: Voting[] } = $props();
</script>

{#if votings.length === 0}
	<p class="empty">No registra votaciones recientes.</p>
{:else}
	<ul>
		{#each votings as voting (voting.votingId)}
			<li><VotingCard {voting} /></li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 22rem), 1fr));
		gap: var(--space-3);
	}

	.empty {
		color: var(--color-text-muted);
	}
</style>
