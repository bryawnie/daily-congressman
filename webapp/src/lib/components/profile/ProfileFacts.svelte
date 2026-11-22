<script lang="ts">
	import type { Profile } from '#lib/types.ts';
	import { ageOn } from '#lib/format/date.ts';
	import { chamberName, constituencyTitle } from '#lib/format/person.ts';

	let { profile, date }: { profile: Profile; date: string } = $props();

	const age = $derived(ageOn(profile.birthdate, date));
	const periods = $derived(profile.periods.map((period) => `${period.from}–${period.to}`).join(', '));
	const constituency = $derived(
		[profile.constituency.name, profile.constituency.detail].filter(Boolean).join(' · ')
	);

	const facts = $derived(
		[
			{ label: 'Cámara', value: chamberName(profile.chamber) },
			{ label: constituencyTitle(profile.constituency.kind), value: constituency },
			{ label: 'Partido', value: profile.party },
			{ label: 'Profesión', value: profile.profession },
			{ label: 'Edad', value: age === null ? '' : `${age} años` },
			{ label: 'Períodos', value: periods }
		].filter((fact) => fact.value)
	);
</script>

<dl>
	{#each facts as fact (fact.label)}
		<div>
			<dt>{fact.label}</dt>
			<dd>{fact.value}</dd>
		</div>
	{/each}
</dl>

<style>
	dl {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 14rem), 1fr));
		gap: var(--space-3) var(--space-5);
		margin: 0;
	}

	dt {
		color: var(--color-text-muted);
		font-size: 0.8125rem;
	}

	dd {
		margin: 0;
	}
</style>
