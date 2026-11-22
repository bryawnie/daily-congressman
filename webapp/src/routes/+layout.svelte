<script lang="ts">
	import SiteFooter from '#lib/components/layout/SiteFooter.svelte';
	import SiteHeader from '#lib/components/layout/SiteHeader.svelte';
	import '#lib/styles/global.css';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href="/favicon.svg" type="image/svg+xml" />
  <link rel="icon" href="/favicon.png" />
</svelte:head>

<SiteHeader />
<main>
	{@render children()}
</main>
<SiteFooter />

<style>
	main {
		max-inline-size: var(--page-width);
		margin-inline: auto;
		padding: var(--space-5) var(--page-gutter) 0;
	}
</style>
