<script lang="ts">
	import { faInstagram, faXTwitter } from '@fortawesome/free-brands-svg-icons';
	import Icon from '#lib/components/ui/Icon.svelte';
	import type { Profile } from '#lib/types.ts';

	let { profile }: { profile: Profile } = $props();

	const links = $derived(
		[
			{ network: 'X', icon: faXTwitter, username: profile.twitterUsername, baseUrl: 'https://x.com/' },
			{ network: 'Instagram', icon: faInstagram, username: profile.instagramUsername, baseUrl: 'https://www.instagram.com/' }
		].filter((link) => link.username)
	);
</script>

{#if links.length > 0}
	<ul>
		{#each links as link (link.network)}
			<li>
				<a href="{link.baseUrl}{link.username}" rel="noopener noreferrer" target="_blank">
					<Icon icon={link.icon} />
					<span class="visually-hidden">{link.network}:</span> @{link.username}
				</a>
			</li>
		{/each}
	</ul>
{/if}

<style>
	a {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}

	.visually-hidden {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	ul {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2) var(--space-4);
		font-size: 0.9375rem;
	}
</style>
