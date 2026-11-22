<script lang="ts">
	import type { MonthTotalsChartProps } from '#lib/types.ts';
	import { formatMonth, formatShortMonth } from '#lib/format/date.ts';
	import { formatCompactPesos, formatPesos } from '#lib/format/money.ts';
	import { ratioOf } from '#lib/format/number.ts';

	let { months, selectedPosition, onSelect }: MonthTotalsChartProps = $props();

	const largestTotal = $derived(Math.max(...months.map((month) => month.total), 0));
</script>

<div class="chart">
	<ul>
		{#each months as month, position (`${month.year}-${month.month}`)}
			<li>
				<button
					type="button"
					aria-pressed={position === selectedPosition}
					aria-label="{formatMonth(month.year, month.month)}: {formatPesos(month.total)}"
					onclick={() => onSelect(position)}
				>
					<span class="total" aria-hidden="true">{formatCompactPesos(month.total)}</span>
					<span class="plot" aria-hidden="true">
						<span class="column" style:block-size="{ratioOf(month.total, largestTotal) * 100}%"></span>
					</span>
					<span class="month" aria-hidden="true">{formatShortMonth(month.month)}</span>
				</button>
			</li>
		{/each}
	</ul>
</div>

<style>
	.chart {
		container-type: inline-size;
	}

	ul {
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: minmax(0, 1fr);
		gap: var(--space-2);
	}

	button {
		display: grid;
		grid-template-rows: auto 8rem auto;
		gap: var(--space-1);
		inline-size: 100%;
		padding: var(--space-2);
		border: 1px solid transparent;
		border-radius: var(--radius-small);
		background: none;
		color: inherit;
		font: inherit;
		text-align: center;
		cursor: pointer;
	}

	button:hover {
		background: var(--color-surface-sunken);
	}

	button[aria-pressed='true'] {
		border-color: var(--color-accent);
		background: color-mix(in oklch, var(--color-accent) 10%, transparent);
	}

	.total {
		font-size: 0.8125rem;
		font-weight: 600;
	}

	.plot {
		display: flex;
		align-items: end;
		justify-content: center;
	}

	.column {
		inline-size: min(100%, 2.5rem);
		min-block-size: 2px;
		border-radius: 0.25rem 0.25rem 0 0;
		background: var(--color-accent);
		opacity: 0.55;
	}

	button[aria-pressed='true'] .column {
		opacity: 1;
	}

	.month {
		color: var(--color-text-muted);
		font-size: 0.8125rem;
	}

	@container (max-width: 30rem) {
		ul {
			gap: 0.125rem;
		}

		button {
			padding: var(--space-1) 0.125rem;
		}

		.total,
		.month {
			font-size: 0.6875rem;
		}
	}
</style>
