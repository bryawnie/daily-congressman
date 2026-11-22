<script lang="ts">
	import type { Profile } from '#lib/types.ts';

	let { profile }: { profile: Profile } = $props();

	const links = $derived(
		[
			{ network: 'X', username: profile.twitterUsername, baseUrl: 'https://x.com/' },
			{ network: 'Instagram', username: profile.instagramUsername, baseUrl: 'https://www.instagram.com/' }
		].filter((link) => link.username)
	);
</script>

{#if links.length > 0}
	<ul>
		{#each links as link (link.network)}
			<li>
				<a href="{link.baseUrl}{link.username}" rel="noopener noreferrer" target="_blank">
					{link.network}: @{link.username}
				</a>
			</li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2) var(--space-4);
		font-size: 0.9375rem;
	}
</style>
