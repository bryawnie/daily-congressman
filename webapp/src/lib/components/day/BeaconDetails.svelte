<script lang="ts">
	import type { Beacon } from '#lib/types.ts';

	let { beacon }: { beacon: Beacon } = $props();

	const facts = $derived([
		{ label: 'Cadena', value: beacon.chainId },
		{ label: 'Pulso', value: beacon.pulseId },
		{ label: 'Valor aleatorio', value: beacon.randOut }
	]);

	const pulseUrl = $derived(
		`https://random.uchile.cl/beacon/2.1-beta/chain/${encodeURIComponent(beacon.chainId)}/pulse/${encodeURIComponent(beacon.pulseId)}`
	);
</script>

<details>
	<summary>Cómo se eligió a este parlamentario</summary>
	<p>
		El sorteo usa el valor aleatorio del pulso de las 00:00 (hora de Chile) del
		<a href="https://random.uchile.cl" target="_blank" rel="noreferrer"
			>faro de aleatoriedad de la Universidad de Chile</a
		>, la lista vigente de parlamentarios y un algoritmo documentado. Con esos tres datos públicos cualquiera puede repetir el cálculo y obtener el mismo
		resultado.
	</p>
	<dl>
		{#each facts as fact (fact.label)}
			<div>
				<dt>{fact.label}</dt>
				<dd>{fact.value}</dd>
			</div>
		{/each}
	</dl>
	<p>
		<a href={pulseUrl} target="_blank" rel="noreferrer">Ver este pulso en el faro de aleatoriedad</a>
	</p>
</details>

<style>
	details {
		padding: var(--space-4) var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface);
	}

	summary {
		cursor: pointer;
		font-weight: 600;
	}

	p {
		margin-block: var(--space-3);
		color: var(--color-text-muted);
		font-size: 0.9375rem;
	}

	dl {
		display: grid;
		gap: var(--space-3);
		margin: 0;
	}

	dt {
		color: var(--color-text-muted);
		font-size: 0.8125rem;
	}

	dd {
		margin: 0;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.8125rem;
		overflow-wrap: anywhere;
	}
</style>
