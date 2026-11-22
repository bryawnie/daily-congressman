<script lang="ts">
	import type { ExpenseExplorerProps } from '#lib/types.ts';
	import { monthSortKey } from '#lib/format/date.ts';
	import MonthBreakdown from './MonthBreakdown.svelte';
	import MonthTotalsChart from './MonthTotalsChart.svelte';

	let { months, chamber }: ExpenseExplorerProps = $props();

	const chronologicalMonths = $derived(
		[...months].sort(
			(first, second) =>
				monthSortKey(first.year, first.month) - monthSortKey(second.year, second.month)
		)
	);
	let selectedPosition = $state<number | null>(null);
	const activePosition = $derived(selectedPosition ?? chronologicalMonths.length - 1);
</script>

{#if chronologicalMonths.length === 0}
	<p class="empty">No registra gastos en los últimos meses.</p>
{:else}
	<MonthTotalsChart
		months={chronologicalMonths}
		selectedPosition={activePosition}
		onSelect={(position) => (selectedPosition = position)}
	/>
	<MonthBreakdown month={chronologicalMonths[activePosition]} {chamber} />
{/if}

<style>
	.empty {
		color: var(--color-text-muted);
	}
</style>
