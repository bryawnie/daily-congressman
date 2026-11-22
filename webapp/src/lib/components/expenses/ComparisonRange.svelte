<script lang="ts">
	import type { ComparisonRangeProps } from '#lib/types.ts';
	import { differenceFromAverage, positionOnScale } from '#lib/expenses/comparison.ts';
	import { formatCompactPesos, formatPesos } from '#lib/format/money.ts';

	let { label, amount, peers }: ComparisonRangeProps = $props();

	const averagePosition = $derived(positionOnScale(peers.avg, peers.min, peers.max));
	const amountPosition = $derived(positionOnScale(amount, peers.min, peers.max));
	const scaleValues = $derived([
		{ name: 'Mínimo', value: peers.min },
		{ name: 'Promedio', value: peers.avg },
		{ name: 'Máximo', value: peers.max }
	]);
</script>

<div class="comparison">
	<div class="heading">
		<h4>{label}</h4>
		<p>
			{differenceFromAverage(amount, peers.avg)}
			{#if peers.ranking}· N.º {peers.ranking} de {peers.count}{/if}
		</p>
	</div>

	<div class="scale" aria-hidden="true">
		<span class="track"></span>
		<span class="marker average" style:inset-inline-start="{averagePosition}%"></span>
		<span class="marker amount" style:inset-inline-start="{amountPosition}%"></span>
	</div>

	<dl class="values">
		{#each scaleValues as scaleValue (scaleValue.name)}
			<div>
				<dt>{scaleValue.name}</dt>
				<dd title={formatPesos(scaleValue.value)}>{formatCompactPesos(scaleValue.value)}</dd>
			</div>
		{/each}
	</dl>
</div>

<style>
	.comparison {
		display: grid;
		gap: var(--space-2);
		padding: var(--space-3) var(--space-4);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-small);
		background: var(--color-surface);
	}

	.heading {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: var(--space-1) var(--space-3);
	}

	h4 {
		font-size: 0.9375rem;
	}

	.heading p {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	.scale {
		position: relative;
		block-size: 1.25rem;
		margin-inline: 0.625rem;
	}

	.track {
		position: absolute;
		inset-block: calc(50% - 0.25rem) auto;
		inset-inline: 0;
		block-size: 0.5rem;
		border-radius: 999px;
		background: var(--color-surface-sunken);
		box-shadow: inset 0 0 0 1px var(--color-border);
	}

	.marker {
		position: absolute;
		translate: -50% 0;
	}

	.average {
		inset-block: 0;
		inline-size: 0.1875rem;
		border-radius: 999px;
		background: var(--color-text-muted);
	}

	.amount {
		inset-block: calc(50% - 0.5rem) auto;
		inline-size: 1rem;
		block-size: 1rem;
		border: 2px solid var(--color-surface);
		border-radius: 50%;
		background: var(--color-accent);
	}

	.values {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
		margin: 0;
		font-size: 0.8125rem;
	}

	.values div {
		display: grid;
	}

	.values div:nth-child(2) {
		text-align: center;
	}

	.values div:last-child {
		text-align: end;
	}

	dt {
		color: var(--color-text-muted);
	}

	dd {
		margin: 0;
		font-weight: 600;
	}
</style>
