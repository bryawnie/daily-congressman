<script lang="ts">
	import DayPage from '#lib/components/day/DayPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if data.day}
	<DayPage day={data.day} daySummaries={data.daySummaries} />
{:else}
	<p>Todavía no se ha sorteado ningún parlamentario. Vuelve después de las 00:00.</p>
{/if}
