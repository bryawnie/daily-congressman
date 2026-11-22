<script lang="ts">
	import DayPage from '#lib/components/day/DayPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<DayPage day={data.day} daySummaries={data.daySummaries} />
