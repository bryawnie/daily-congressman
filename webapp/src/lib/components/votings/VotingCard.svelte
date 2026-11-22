<script lang="ts">
	import type { Voting } from '#lib/types.ts';
	import Badge from '#lib/components/ui/Badge.svelte';
	import { formatMediumDate } from '#lib/format/date.ts';
	import { resultTone, voteTone } from '#lib/votings/tones.ts';
	import VoteTally from './VoteTally.svelte';

	let { voting }: { voting: Voting } = $props();
</script>

<article>
	<header>
		<h3>{voting.documentTitle || 'Votación sin título'}</h3>
		<p class="meta">
			{#if voting.bulletinNumber}Boletín {voting.bulletinNumber} · {/if}
			<time datetime={voting.votingDate}>{formatMediumDate(voting.votingDate)}</time>
		</p>
	</header>

	<dl class="verdicts">
		<div>
			<dt>Su voto</dt>
			<dd><Badge tone={voteTone(voting.voted)}>{voting.voted}</Badge></dd>
		</div>
		<div>
			<dt>Resultado</dt>
			<dd><Badge tone={resultTone(voting.result)}>{voting.result || 'Sin resultado'}</Badge></dd>
		</div>
	</dl>

	<VoteTally {voting} />

	{#if voting.articleText}
		<details>
			<summary>Texto votado</summary>
			<p>{voting.articleText}</p>
		</details>
	{/if}
</article>

<style>
	article {
		display: grid;
		gap: var(--space-3);
		padding: var(--space-4);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-small);
		background: var(--color-surface-sunken);
	}

	h3 {
		font-size: 1rem;
	}

	.meta {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	.verdicts {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2) var(--space-5);
		margin: 0;
	}

	.verdicts div {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	dt {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	dd {
		margin: 0;
	}

	summary {
		cursor: pointer;
		color: var(--color-accent);
		font-size: 0.875rem;
	}

	details p {
		margin-block-start: var(--space-2);
		font-size: 0.9375rem;
	}
</style>
