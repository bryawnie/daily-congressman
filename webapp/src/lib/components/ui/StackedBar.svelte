<script lang="ts">
	import type { ToneSegment } from '#lib/types.ts';
	import { wipe } from '#lib/motion/actions.ts';
	import { formatPercent, ratioOf } from '#lib/format/number.ts';

	let { segments }: { segments: ToneSegment[] } = $props();

	const total = $derived(segments.reduce((sum, segment) => sum + segment.value, 0));
</script>

<div class="stack">
	<div class="bar" aria-hidden="true" use:wipe>
		{#each segments as segment (segment.label)}
			{#if segment.value > 0}
				<span
					class="segment"
					data-tone={segment.tone}
					style:inline-size="{ratioOf(segment.value, total) * 100}%"
				></span>
			{/if}
		{/each}
	</div>
	<ul class="legend">
		{#each segments as segment (segment.label)}
			<li data-tone={segment.tone}>
				<span class="swatch" aria-hidden="true"></span>
				<span>{segment.label}</span>
				<strong>{segment.value}</strong>
				<span class="share">({formatPercent(ratioOf(segment.value, total))})</span>
			</li>
		{/each}
	</ul>
</div>

<style>
	.stack {
		display: grid;
		gap: var(--space-3);
	}

	.bar {
		display: flex;
		block-size: 0.875rem;
		overflow: hidden;
		border-radius: 999px;
		background: var(--color-surface-sunken);
	}

	.segment {
		background: var(--tone);
	}

	.segment + .segment {
		border-inline-start: 2px solid var(--color-surface);
	}

	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2) var(--space-5);
		font-size: 0.875rem;
	}

	.legend li {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.swatch {
		inline-size: 0.75rem;
		block-size: 0.75rem;
		border-radius: 0.1875rem;
		background: var(--tone);
	}

	.share {
		color: var(--color-text-muted);
	}
</style>
