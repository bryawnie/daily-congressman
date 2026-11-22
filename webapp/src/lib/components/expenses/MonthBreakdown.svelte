<script lang="ts">
	import type { MonthBreakdownProps } from '#lib/types.ts';
	import { formatMonth } from '#lib/format/date.ts';
	import { formatPesos } from '#lib/format/money.ts';
	import ComparisonKey from './ComparisonKey.svelte';
	import ExpenseTypePanel from './ExpenseTypePanel.svelte';

	let { month, chamber }: MonthBreakdownProps = $props();
</script>

<div class="breakdown">
	<header>
		<h3>{formatMonth(month.year, month.month)}</h3>
		<p>Total del mes: <strong>{formatPesos(month.total)}</strong></p>
	</header>
	<ComparisonKey />
	<div class="panels">
		{#each month.detail as detail (detail.type)}
			<ExpenseTypePanel {detail} {chamber} />
		{/each}
	</div>
</div>

<style>
	.breakdown {
		display: grid;
		gap: var(--space-4);
	}

	header {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-1) var(--space-4);
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
