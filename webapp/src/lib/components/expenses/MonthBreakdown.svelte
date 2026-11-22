<script lang="ts">
	import type { MonthBreakdownProps } from '#lib/types.ts';
	import { formatMonth } from '#lib/format/date.ts';
	import ComparisonKey from './ComparisonKey.svelte';
	import ExpenseTypePanel from './ExpenseTypePanel.svelte';

	let { month, chamber }: MonthBreakdownProps = $props();
</script>

<section class="breakdown">
	<h3>Detalle de {formatMonth(month.year, month.month)}</h3>
	<ComparisonKey />
	<div class="panels">
		{#each month.detail as detail (detail.type)}
			<ExpenseTypePanel {detail} {chamber} />
		{/each}
	</div>
</section>

<style>
	.breakdown {
		display: grid;
		gap: var(--space-4);
		padding: var(--space-4) var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface-sunken);
	}

	h3 {
		font-size: 1.0625rem;
	}

	.panels {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 24rem), 1fr));
		gap: var(--space-4);
		align-items: start;
	}
</style>
