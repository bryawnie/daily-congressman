<script lang="ts">
	import { faArrowUpRightFromSquare } from '@fortawesome/free-solid-svg-icons';
	import type { SectionCardProps } from '#lib/types.ts';
	import Icon from './Icon.svelte';

	let { title, icon, description, source = null, pending = false, children }: SectionCardProps = $props();
	const headingId = $props.id();
</script>

<section class="card" aria-labelledby={headingId}>
	<header>
		<h2 id={headingId}>
			{#if icon}<Icon {icon} />{/if}
			{title}
		</h2>
		{#if description}
			<p class="description">{description}</p>
		{/if}
	</header>
	{#if pending}
		<p class="pending">Estos datos todavía se están recopilando. Vuelve a revisar más tarde.</p>
	{:else}
		{@render children()}
	{/if}
	{#if source}
		<p class="source">
			<a href={source.href} target="_blank" rel="noreferrer">{source.label}
				<Icon icon={faArrowUpRightFromSquare} /></a
			>
		</p>
	{/if}
</section>

<style>
	.card {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		padding: var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface);
		min-inline-size: 0;
	}

	h2 {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		font-size: 1.125rem;
	}

	h2 :global(.icon) {
		color: var(--color-accent);
	}

	.description,
	.pending {
		color: var(--color-text-muted);
		font-size: 0.9375rem;
	}

	.source {
		margin-block-start: auto;
		text-align: end;
		font-size: 0.75rem;
	}

	header {
		display: grid;
		gap: var(--space-1);
	}
</style>
