<script lang="ts">
	import type { ExpenseTypePanelProps } from '#lib/types.ts';
	import { formatPesos } from '#lib/format/money.ts';
	import CategoryBreakdown from './CategoryBreakdown.svelte';
	import ComparisonList from './ComparisonList.svelte';

	let { detail, chamber }: ExpenseTypePanelProps = $props();
</script>

<section>
	<header>
		<h3>{detail.type}</h3>
		{#if detail.amount === null}
			<p class="amount">Sin gasto registrado</p>
		{:else}
			<p class="amount">{formatPesos(detail.amount)}</p>
		{/if}
		{#if detail.supportStaff}
			<p class="note">{detail.supportStaff} personas de apoyo contratadas</p>
		{/if}
	</header>

	{#if detail.amount !== null}
		<div class="comparisons">
			<ComparisonList amount={detail.amount} comparison={detail.stats} {chamber} />
		</div>
	{/if}

	{#if detail.expenses?.length}
		<div class="categories">
			<h4>Por categoría</h4>
			<CategoryBreakdown categories={detail.expenses} {chamber} />
		</div>
	{/if}
</section>

<style>
	section {
		display: grid;
		gap: var(--space-4);
		padding: var(--space-4);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-small);
	}

	header {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-1) var(--space-4);
	}

	h3 {
		font-size: 1rem;
	}

	.amount {
		font-size: 1.5rem;
		font-weight: 700;
	}

	.note {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	.comparisons,
	.categories {
		display: grid;
		gap: var(--space-4);
	}

	h4 {
		font-size: 0.9375rem;
	}
</style>
