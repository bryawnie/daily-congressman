<script lang="ts">
	import type { MonthSummaryProps } from '#lib/types.ts';
	import { formatMonth } from '#lib/format/date.ts';
	import { formatPesos } from '#lib/format/money.ts';

	let { month }: MonthSummaryProps = $props();

	const types = $derived(month.detail.filter((detail) => detail.amount !== null));
</script>

<section>
	<h3>Resumen de {formatMonth(month.year, month.month)}</h3>
	<p class="total">{formatPesos(month.total)}</p>
	<p class="label">Total del mes</p>
	{#if types.length > 1}
		<dl>
			{#each types as detail (detail.type)}
				<div>
					<dt>{detail.type}</dt>
					<dd>{formatPesos(detail.amount ?? 0)}</dd>
				</div>
			{/each}
		</dl>
	{/if}
</section>

<style>
	section {
		display: grid;
		gap: var(--space-1);
		padding: var(--space-4) var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface-sunken);
	}

	h3 {
		font-size: 1.0625rem;
	}

	.total {
		font-size: 1.75rem;
		font-weight: 700;
	}

	.label {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	dl {
		display: grid;
		gap: var(--space-2);
		margin: var(--space-3) 0 0;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
	}

	dt {
		color: var(--color-text-muted);
	}

	dd {
		margin: 0;
		font-weight: 600;
	}
</style>
