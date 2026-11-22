<script lang="ts">
	import type { DayPickerProps } from '#lib/types.ts';
	import { formatShortDate } from '#lib/format/date.ts';

	let { days, currentDate }: DayPickerProps = $props();
</script>

{#if days.length > 0}
	<nav aria-label="Días recientes">
		<ul>
			{#each days as day (day.date)}
				<li>
					<a
						href="/dias/{day.date}"
						aria-current={day.date === currentDate ? 'page' : undefined}
					>
						<time datetime={day.date}>{formatShortDate(day.date)}</time>
						<span class="name">{day.fullName}</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>
{/if}

<style>
	ul {
		display: flex;
		gap: var(--space-2);
		overflow-x: auto;
		padding-block-end: var(--space-2);
		scroll-snap-type: x proximity;
	}

	li {
		flex: 0 0 auto;
		scroll-snap-align: start;
	}

	a {
		display: grid;
		gap: 0.125rem;
		inline-size: 10rem;
		padding: var(--space-2) var(--space-3);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-small);
		background: var(--color-surface);
		color: var(--color-text);
		text-decoration: none;
	}

	a:hover {
		border-color: var(--color-accent);
	}

	a[aria-current='page'] {
		border-color: var(--color-accent);
		background: color-mix(in oklch, var(--color-accent) 10%, var(--color-surface));
	}

	time {
		font-size: 0.875rem;
		font-weight: 600;
		text-transform: capitalize;
	}

	.name {
		overflow: hidden;
		color: var(--color-text-muted);
		font-size: 0.8125rem;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
