<script lang="ts">
	import type { BarItem } from '#lib/types.ts';
	import { ratioOf } from '#lib/format/number.ts';
	import ProportionBar from './ProportionBar.svelte';

	let { items }: { items: BarItem[] } = $props();

	const largestValue = $derived(Math.max(...items.map((item) => item.value), 0));
</script>

<ul class="bars">
	{#each items as item (item.label)}
		<li>
			<span>{item.label}</span>
			<strong>{item.value}</strong>
			<div class="bar"><ProportionBar ratio={ratioOf(item.value, largestValue)} /></div>
		</li>
	{/each}
</ul>

<style>
	.bars {
		display: grid;
		gap: var(--space-3);
	}

	li {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: var(--space-1) var(--space-3);
		font-size: 0.9375rem;
	}

	.bar {
		grid-column: 1 / -1;
	}
</style>
