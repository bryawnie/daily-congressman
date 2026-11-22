<script lang="ts">
	import { page } from '$app/state';

	const isMissingPage = $derived(page.status === 404);
</script>

<svelte:head>
	<title>{isMissingPage ? 'Página no encontrada' : 'Algo salió mal'} · Congresista del día</title>
</svelte:head>

<section>
	<p class="status">{page.status}</p>
	<h1>{isMissingPage ? 'No encontramos esta página' : 'Algo salió mal'}</h1>
	<p class="explanation">
		{#if isMissingPage}
			Solo se muestran los últimos días sorteados. Puede que esta fecha no se haya sorteado todavía.
		{:else}
			No pudimos mostrar esta página. Intenta de nuevo más tarde.
		{/if}
	</p>
	<a href="/">Ver el congresista del día</a>
</section>

<style>
	section {
		display: grid;
		justify-items: start;
		gap: var(--space-3);
		padding-block: var(--space-6);
	}

	.status {
		color: var(--color-text-muted);
		font-size: 0.875rem;
		font-weight: 600;
	}

	h1 {
		font-size: clamp(1.5rem, 4vw, 2rem);
	}

	.explanation {
		color: var(--color-text-muted);
		max-inline-size: 36rem;
	}
</style>
