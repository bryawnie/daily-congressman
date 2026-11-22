<script lang="ts">
	import type { ComparisonListProps } from '#lib/types.ts';
	import { peerComparisonsOf } from '#lib/expenses/comparison.ts';
	import ComparisonRange from './ComparisonRange.svelte';

	let { amount, comparison, chamber }: ComparisonListProps = $props();

	const comparisons = $derived(peerComparisonsOf(comparison, chamber));
</script>

{#each comparisons as { label, peers } (label)}
	<ComparisonRange {label} {amount} {peers} />
{/each}
