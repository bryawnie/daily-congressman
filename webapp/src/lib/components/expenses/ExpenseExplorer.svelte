<script lang="ts">
	import type { ExpenseExplorerProps } from '#lib/types.ts';
	import { monthSortKey } from '#lib/format/date.ts';
	import { faArrowUpRightFromSquare } from '@fortawesome/free-solid-svg-icons';
	import Icon from '#lib/components/ui/Icon.svelte';
	import MonthSummary from './MonthSummary.svelte';
	import MonthBreakdown from './MonthBreakdown.svelte';
	import MonthTotalsChart from './MonthTotalsChart.svelte';

	let { months, chamber, source = null }: ExpenseExplorerProps = $props();

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
	{#if source}
		<p class="source">
			<a href={source.href} target="_blank" rel="noreferrer">{source.label}
				<Icon icon={faArrowUpRightFromSquare} /></a
			>
		</p>
	{/if}
	<MonthSummary month={chronologicalMonths[activePosition]} />
	<MonthBreakdown month={chronologicalMonths[activePosition]} {chamber} />
{/if}

<style>
	.source {
		text-align: end;
		font-size: 0.75rem;
	}

	.empty {
		color: var(--color-text-muted);
	}
</style>
