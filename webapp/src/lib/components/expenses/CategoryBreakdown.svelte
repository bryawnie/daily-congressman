<script lang="ts">
	import type { CategoryBreakdownProps } from '#lib/types.ts';
	import CategoryRow from './CategoryRow.svelte';

	let { categories, chamber }: CategoryBreakdownProps = $props();

	const categoriesByAmount = $derived([...categories].sort((first, second) => second.amount - first.amount));
	const largestAmount = $derived(categoriesByAmount[0]?.amount ?? 0);
</script>

<ul>
	{#each categoriesByAmount as category (category.subtype)}
		<li><CategoryRow {category} {largestAmount} {chamber} /></li>
	{/each}
</ul>

<style>
	ul {
		display: grid;
		gap: var(--space-4);
	}
</style>
