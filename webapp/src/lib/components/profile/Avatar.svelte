<script lang="ts">
	import type { Profile } from '#lib/types.ts';
	import { fullName, initialsOf } from '#lib/format/person.ts';

	let { profile }: { profile: Profile } = $props();
</script>

{#if profile.picture}
	<img
		class="avatar"
		src={profile.picture}
		alt="Fotografía de {fullName(profile)}"
		width="144"
		height="192"
		referrerpolicy="no-referrer"
	/>
{:else}
	<span class="avatar placeholder" aria-hidden="true">{initialsOf(profile)}</span>
{/if}

<style>
	.avatar {
		inline-size: 9rem;
		aspect-ratio: 3 / 4;
		border-radius: var(--radius-medium);
		object-fit: cover;
		object-position: center top;
		background: var(--color-surface-sunken);
	}

	.placeholder {
		display: grid;
		place-items: center;
		color: var(--color-text-muted);
		font-size: 2.5rem;
		font-weight: 700;
	}
</style>
