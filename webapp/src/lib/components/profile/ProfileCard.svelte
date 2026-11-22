<script lang="ts">
	import type { Profile } from '#lib/types.ts';
	import { formatLongDate } from '#lib/format/date.ts';
	import { fullName, roleTitle } from '#lib/format/person.ts';
	import Avatar from './Avatar.svelte';
	import ProfileFacts from './ProfileFacts.svelte';
	import SocialLinks from './SocialLinks.svelte';

	let { profile, date }: { profile: Profile; date: string } = $props();
</script>

<article>
	<div class="identity">
		<Avatar {profile} />
		<div class="heading">
			<p class="eyebrow">
				{roleTitle(profile)} del día · <time datetime={date}>{formatLongDate(date)}</time>
			</p>
			<h1>{fullName(profile)}</h1>
			{#if profile.partyAlias}
				<p class="party">{profile.partyAlias}</p>
			{/if}
			<SocialLinks {profile} />
		</div>
	</div>
	<ProfileFacts {profile} {date} />
</article>

<style>
	article {
		display: grid;
		gap: var(--space-5);
		padding: var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface);
	}

	.identity {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-5);
		align-items: center;
	}

	.heading {
		display: grid;
		gap: var(--space-2);
		flex: 1 1 18rem;
	}

	.eyebrow {
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	h1 {
		font-size: clamp(1.75rem, 5vw, 2.5rem);
	}

	.party {
		font-size: 1.0625rem;
	}
</style>
