<script lang="ts">
	import type { Attendance, ToneSegment } from '#lib/types.ts';
	import { formatPercent, ratioOf } from '#lib/format/number.ts';
	import StackedBar from '#lib/components/ui/StackedBar.svelte';

	let { attendance }: { attendance: Attendance } = $props();

	const unaccountedSessions = $derived(
		Math.max(
			0,
			attendance.total -
				attendance.attended -
				attendance.justifiedAbsent -
				attendance.unjustifiedAbsent
		)
	);

	const segments: ToneSegment[] = $derived([
		{ label: 'Asistió', value: attendance.attended, tone: 'positive' },
		{ label: 'Ausencia justificada', value: attendance.justifiedAbsent, tone: 'warning' },
		{ label: 'Ausencia injustificada', value: attendance.unjustifiedAbsent, tone: 'negative' },
		...(unaccountedSessions > 0
			? [{ label: 'Sin información', value: unaccountedSessions, tone: 'neutral' as const }]
			: [])
	]);
</script>

{#if attendance.total === 0}
	<p class="empty">Todavía no hay sesiones registradas.</p>
{:else}
	<p class="headline">
		<strong>{formatPercent(ratioOf(attendance.attended, attendance.total))}</strong>
		<span>asistió a {attendance.attended} de {attendance.total} sesiones</span>
	</p>
	<StackedBar {segments} />
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
