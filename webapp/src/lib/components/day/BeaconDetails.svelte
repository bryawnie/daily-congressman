<script lang="ts">
	import {
		faArrowUpRightFromSquare,
		faCheck,
		faCopy,
		faFingerprint,
		faTerminal
	} from '@fortawesome/free-solid-svg-icons';
	import Icon from '#lib/components/ui/Icon.svelte';
	import type { Beacon } from '#lib/types.ts';

	let { beacon }: { beacon: Beacon } = $props();

	const facts = $derived([
		{ label: 'Cadena', value: beacon.chainId },
		{ label: 'Pulso', value: beacon.pulseId },
		{ label: 'Valor aleatorio', value: beacon.randOut, copyable: true }
	]);

	let copied = $state(false);

	async function copyRandomValue() {
		try {
			await navigator.clipboard.writeText(beacon.randOut);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			copied = false;
		}
	}

	const pulseUrl = $derived(
		`https://random.uchile.cl/beacon/2.1-beta/chain/${encodeURIComponent(beacon.chainId)}/pulse/${encodeURIComponent(beacon.pulseId)}`
	);
</script>

<section>
	<h2><Icon icon={faFingerprint} /> Cómo se eligió a este parlamentario</h2>
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
				<dd>
					<span>{fact.value}</span>
					{#if fact.copyable}
						<button type="button" class="copy" onclick={copyRandomValue}>
							<Icon icon={copied ? faCheck : faCopy} />
							{copied ? 'Copiado' : 'Copiar'}
						</button>
					{/if}
				</dd>
			</div>
		{/each}
	</dl>
	<div class="actions">
		<a class="button primary" href={pulseUrl} target="_blank" rel="noreferrer">
			Ver pulso
			<Icon icon={faArrowUpRightFromSquare} />
		</a>
		<a class="button" href="/como-funciona">
			<Icon icon={faTerminal} />
			Quiero verificar el resultado
		</a>
	</div>
</section>

<style>
	section {
		padding: var(--space-4) var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-medium);
		background: var(--color-surface);
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
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-1) var(--space-3);
		margin: 0;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.8125rem;
	}

	dd span {
		overflow-wrap: anywhere;
	}

	.copy {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		padding: 0.125rem var(--space-2);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-small);
		background: var(--color-surface-sunken);
		color: var(--color-accent);
		font: inherit;
		font-family: system-ui, sans-serif;
		font-size: 0.75rem;
		cursor: pointer;
	}

	.copy:hover {
		border-color: var(--color-accent);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-3);
		margin-block-start: var(--space-4);
	}

	.button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-4);
		border: 1px solid var(--color-accent);
		border-radius: 999px;
		background: color-mix(in oklch, var(--color-accent) 12%, transparent);
		font-size: 0.875rem;
		font-weight: 600;
		text-decoration: none;
		transition: background-color 0.2s ease;
	}

	.button:hover {
		background: color-mix(in oklch, var(--color-accent) 24%, transparent);
	}

	.button.primary {
		background: var(--color-accent);
		color: var(--color-surface);
	}

	.button.primary:hover {
		background: color-mix(in oklch, var(--color-accent) 85%, var(--color-text));
	}
</style>
