<script lang="ts">
	import type { ToneSegment, Voting } from '#lib/types.ts';
	import StackedBar from '#lib/components/ui/StackedBar.svelte';

	let { voting }: { voting: Voting } = $props();

	const segments: ToneSegment[] = $derived([
		{ label: 'A favor', value: voting.totalApproved, tone: 'positive' },
		{ label: 'En contra', value: voting.totalRejected, tone: 'negative' },
		{ label: 'Abstención', value: voting.totalAbstention, tone: 'warning' }
	]);
</script>

<StackedBar {segments} />
{#if voting.quorum}
	<p class="quorum">Quórum exigido: {voting.quorum}</p>
{/if}

<style>
	.quorum {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}
</style>
