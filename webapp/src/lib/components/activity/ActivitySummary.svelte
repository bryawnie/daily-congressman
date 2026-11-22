<script lang="ts">
	import type { Activity } from '#lib/types.ts';
	import HorizontalBars from '#lib/components/ui/HorizontalBars.svelte';

	let { activity }: { activity: Activity } = $props();

	const statusBars = $derived(
		Object.entries(activity.byStatus)
			.map(([status, count]) => ({ label: status, value: count }))
			.sort((first, second) => second.value - first.value)
	);
</script>

{#if activity.all === 0}
	<p class="empty">No registra proyectos ni mociones.</p>
{:else}
	<p class="headline">
		<strong>{activity.all}</strong>
		<span>{activity.all === 1 ? 'proyecto presentado' : 'proyectos presentados'}, por estado</span>
	</p>
	<HorizontalBars items={statusBars} />
{/if}

<style>
	.headline {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-2) var(--space-3);
	}

	.headline strong {
		font-size: 2rem;
		line-height: 1;
	}

	.empty {
		color: var(--color-text-muted);
	}
</style>
