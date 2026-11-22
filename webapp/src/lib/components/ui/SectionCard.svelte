<script lang="ts">
	import type { SectionCardProps } from '#lib/types.ts';

	let { title, description, pending = false, children }: SectionCardProps = $props();
	const headingId = $props.id();
</script>

<section class="card" aria-labelledby={headingId}>
	<header>
		<h2 id={headingId}>{title}</h2>
		{#if description}
			<p class="description">{description}</p>
		{/if}
	</header>
	{#if pending}
		<p class="pending">Estos datos todavía se están recopilando. Vuelve a revisar más tarde.</p>
	{:else}
		{@render children()}
	{/if}
</section>

<style>
	.card {
		display: grid;
		gap: var(--space-4);
		align-content: start;
		padding: var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface);
		min-inline-size: 0;
	}

	h2 {
		font-size: 1.125rem;
	}

	.description,
	.pending {
		color: var(--color-text-muted);
		font-size: 0.9375rem;
	}

	header {
		display: grid;
		gap: var(--space-1);
	}
</style>
