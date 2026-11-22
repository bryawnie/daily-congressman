<script lang="ts">
	import type { CategoryRowProps } from '#lib/types.ts';
	import ProportionBar from '#lib/components/ui/ProportionBar.svelte';
	import { peerComparisonsOf } from '#lib/expenses/comparison.ts';
	import { formatPesos } from '#lib/format/money.ts';
	import { ratioOf } from '#lib/format/number.ts';
	import ComparisonList from './ComparisonList.svelte';

	let { category, largestAmount, chamber }: CategoryRowProps = $props();

	const canCompare = $derived(peerComparisonsOf(category.stats, chamber).length > 0);
</script>

<div class="row">
	<div class="summary">
		<span>{category.subtype}</span>
		<strong>{formatPesos(category.amount)}</strong>
	</div>
	<ProportionBar ratio={ratioOf(category.amount, largestAmount)} />
	{#if canCompare}
		<details>
			<summary>Comparar con otros parlamentarios</summary>
			<div class="comparisons">
				<ComparisonList amount={category.amount} comparison={category.stats} {chamber} />
			</div>
		</details>
	{/if}
</div>

<style>
	.row {
		display: grid;
		gap: var(--space-2);
	}

	.summary {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
		font-size: 0.9375rem;
	}

	summary {
		cursor: pointer;
		color: var(--color-accent);
		font-size: 0.8125rem;
	}

	.comparisons {
		display: grid;
		gap: var(--space-4);
		margin-block-start: var(--space-3);
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-small);
		background: var(--color-surface-sunken);
	}
</style>
