<script lang="ts">
	import type { DayPageProps } from '#lib/types.ts';
	import ActivitySummary from '#lib/components/activity/ActivitySummary.svelte';
	import AttendanceSummary from '#lib/components/attendance/AttendanceSummary.svelte';
	import ExpenseExplorer from '#lib/components/expenses/ExpenseExplorer.svelte';
	import ProfileCard from '#lib/components/profile/ProfileCard.svelte';
	import SectionCard from '#lib/components/ui/SectionCard.svelte';
	import VotingList from '#lib/components/votings/VotingList.svelte';
	import { sourceLinksOf } from '#lib/sources/links.ts';
	import { fullName, roleTitle } from '#lib/format/person.ts';
	import BeaconDetails from './BeaconDetails.svelte';
	import DayPicker from './DayPicker.svelte';

	let { day, daySummaries }: DayPageProps = $props();

	const pendingSections = $derived(new Set(day.pending));
	const sources = $derived(sourceLinksOf(day.congressman));
	const name = $derived(fullName(day.congressman));
</script>

<svelte:head>
	<title>{name} · Congresista del día</title>
	<meta
		name="description"
		content="{roleTitle(day.congressman)} {name}: asistencia, votaciones y gastos del congresista del día."
	/>
</svelte:head>

<div class="page">
	<DayPicker days={daySummaries} currentDate={day.date} />
	<ProfileCard profile={day.congressman} date={day.date} />

	<div class="pair">
		<SectionCard title="Asistencia a sesiones" source={sources.attendance} pending={pendingSections.has('attendance')}>
			{#if day.attendance}
				<AttendanceSummary attendance={day.attendance} />
			{/if}
		</SectionCard>
		<SectionCard title="Proyectos presentados" source={sources.activity} pending={pendingSections.has('activity')}>
			{#if day.activity}
				<ActivitySummary activity={day.activity} />
			{/if}
		</SectionCard>
	</div>

	<SectionCard
		title="Últimas votaciones"
		source={sources.votings}
		description="Cómo votó y cómo votó el resto de la sala."
		pending={pendingSections.has('votings')}
	>
		<VotingList votings={day.votings} />
	</SectionCard>

	<SectionCard
		title="Gastos"
		description="Comparados con los demás parlamentarios. El puesto N.º 1 es quien más gastó."
		pending={pendingSections.has('expenses')}
	>
		<ExpenseExplorer
			months={day.expenses}
			chamber={day.congressman.chamber}
			source={sources.expenses}
		/>
	</SectionCard>

	<BeaconDetails beacon={day.beacon} />
</div>

<style>
	.page {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: var(--space-5);
	}

	.pair {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 22rem), 1fr));
		gap: var(--space-5);
	}
</style>
