<script lang="ts">
  import {
    faArrowLeft,
    faCheck,
    faCopy,
    faDownload,
    faTerminal,
  } from "@fortawesome/free-solid-svg-icons";
  import Icon from "#lib/components/ui/Icon.svelte";
  import { tokenizePython } from "#lib/highlight/python.ts";
  import script from "../../../static/verificar_sorteo.py?raw";

  const scriptTokens = tokenizePython(script);
  const beaconUrl = "https://random.uchile.cl";

  const steps = [
    {
      title: "Comprueba el pulso",
      text: "Abre el enlace al pulso y revisa que su valor aleatorio coincide con el que muestra este sitio.",
    },
    {
      title: "Reúne la lista",
      text: "Obtén la lista vigente de parlamentarios, por ejemplo desde el sitio oficial del Senado, y ordénala por identificador.",
    },
    {
      title: "Calcula la posición",
      text: "Calcula R mod N y revisa que la persona en esa posición es la elegida. El script de más abajo lo hace por ti.",
    },
  ];

  let copied = $state(false);

  async function copyScript() {
    try {
      await navigator.clipboard.writeText(script);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      copied = false;
    }
  }
</script>

<svelte:head>
  <title>Cómo funciona el sorteo · Parlamentario del día</title>
  <meta
    name="description"
    content="Cómo se elige al parlamentario del día con aleatoriedad pública y cómo repetir el sorteo tú mismo."
  />
</svelte:head>

<article>
  <p class="back">
    <a href="/"><Icon icon={faArrowLeft} /> Volver al parlamentario del día</a>
  </p>
  <h1>Cómo funciona el sorteo</h1>
  <p class="lead">
    El parlamentario del día es seleccionado mediante un número aleatorio público
    que ninguna persona controla, y cualquiera puede repetir el cálculo para
    comprobarlo.
  </p>

  <section>
    <h2>Qué es la aleatoriedad pública</h2>
    <p>
      Un <strong>faro de aleatoriedad</strong> es un servicio que publica, cada
      cierto tiempo, un número aleatorio al que todos tienen acceso. Cada
      publicación se llama <strong>pulso</strong>. Usamos el faro de la
      Universidad de Chile (<a href={beaconUrl} target="_blank" rel="noreferrer"
        >random.uchile.cl</a
      >). El pulso cumple con lo siguiente:
    </p>
    <ul>
      <li>
        <strong>Imprevisible:</strong> el valor de un pulso no se puede saber antes
        de que se publique, ni siquiera quienes hacen este sitio.
      </li>
      <li>
        <strong>Público:</strong> cualquiera puede consultar el pulso y ver que es
        el mismo que usamos.
      </li>
      <li>
        <strong>Fijo después de publicado:</strong> los pulsos quedan encadenados,
        así que un pulso no se puede cambiar a posteriori sin que se note.
      </li>
    </ul>
    <p>
      Cada pulso se identifica con una <strong>cadena</strong> y un
      <strong>número de pulso</strong>
      y trae un <strong>valor aleatorio</strong>, un número muy grande escrito
      en hexadecimal.
    </p>
  </section>

  <section>
    <h2>Cómo se elige al parlamentario</h2>
    <ol>
      <li>
        Se toma el pulso que el faro generó a las <strong
          >00:00 hora de Chile</strong
        > del día.
      </li>
      <li>
        Se ordena la lista de parlamentarios vigentes por su identificador, de
        menor a mayor.
      </li>
      <li>
        Se lee el valor aleatorio del pulso como un entero grande <code>R</code
        >.
      </li>
      <li>
        Se elige al parlamentario en la posición <code>R mod N</code>, donde
        <code>N</code>
        es la cantidad de parlamentarios (el resto de dividir <code>R</code>
        entre <code>N</code>; la primera posición es la 0).
      </li>
    </ol>
    <p>
      Como todos los datos de entrada son públicos y el cálculo es siempre el
      mismo, cualquiera que los reúna obtiene el mismo resultado.
    </p>
  </section>

  <section>
    <h2>Cómo verificar un día</h2>
    <p>
      En cada día, el recuadro «Cómo se eligió a este parlamentario» muestra la
      cadena, el pulso y el valor aleatorio usados, con un enlace al pulso en el
      sitio del faro. Para repetir el sorteo:
    </p>
    <ol class="steps">
      {#each steps as step (step.title)}
        <li>
          <h3>{step.title}</h3>
          <p>{step.text}</p>
        </li>
      {/each}
    </ol>

    <h3 class="subheading">O hazlo con un script</h3>
    <p>
      Funciona con Python 3.9 o superior y no necesita instalar nada. Hoy el
      sorteo incluye solo al Senado. Ejecútalo con la fecha que quieres
      comprobar:
    </p>

    <div class="terminal">
      <p class="terminal-bar"><Icon icon={faTerminal} /> Terminal</p>
      <pre><code
          ><span class="prompt">$</span> python3 verificar_sorteo.py 2026-11-20
<span class="output">Cristian Vial Maceratta</span></code
        ></pre>
    </div>

    <div class="code">
      <div class="code-bar">
        <span class="filename">verificar_sorteo.py</span>
        <span class="actions">
          <button type="button" onclick={copyScript}>
            <Icon icon={copied ? faCheck : faCopy} />
            {copied ? "Copiado" : "Copiar"}
          </button>
          <a href="/verificar_sorteo.py" download
            ><Icon icon={faDownload} /> Descargar</a
          >
        </span>
      </div>
      <pre><code
          >{#each scriptTokens as token}{#if token.kind === "plain"}{token.text}{:else}<span
                class={token.kind}>{token.text}</span
              >{/if}{/each}</code
        ></pre>
    </div>
  </section>

  <section>
    <h2>Qué tener en cuenta</h2>
    <ul>
      <li>
        El sitio guarda la elección cuando la hace y no la recalcula. El script
        usa la lista de parlamentarios vigente <em>hoy</em>, así que si un
        escaño cambió de ocupante desde el sorteo, un día antiguo podría no
        coincidir.
      </li>
      <li>
        El script toma el primer pulso generado a las 00:00 (hora de Chile) o
        después, igual que el sitio.
      </li>
      <li>
        En macOS, si Python falla con un error de certificado, ejecuta «Install
        Certificates.command» de tu instalación de Python.
      </li>
    </ul>
  </section>
</article>

<style>
  article {
    display: grid;
    gap: var(--space-5);
    max-inline-size: 46rem;
    padding-block-end: var(--space-5);
  }

  .back a {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: 1px solid var(--color-accent);
    border-radius: 999px;
    background: color-mix(in oklch, var(--color-accent) 12%, transparent);
    font-size: 0.9375rem;
    font-weight: 600;
    text-decoration: none;
    transition:
      background-color 0.2s ease,
      transform 0.2s ease;
  }

  .back a:hover {
    background: color-mix(in oklch, var(--color-accent) 24%, transparent);
    transform: translateX(-2px);
  }

  h1 {
    font-size: clamp(1.5rem, 4vw, 2rem);
  }

  .lead {
    color: var(--color-text-muted);
    font-size: 1.0625rem;
  }

  section {
    display: grid;
    gap: var(--space-3);
  }

  h2 {
    font-size: 1.25rem;
  }

  h3 {
    margin-block-start: var(--space-2);
    font-size: 1.0625rem;
  }

  ul,
  ol {
    display: grid;
    gap: var(--space-2);
    margin: 0;
    padding-inline-start: var(--space-5);
  }

  ul {
    list-style: disc;
  }

  ol {
    list-style: decimal;
  }

  code {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.875em;
  }

  .steps {
    counter-reset: step;
    gap: var(--space-3);
    padding-inline-start: 0;
    list-style: none;
  }

  .steps li {
    counter-increment: step;
    position: relative;
    display: grid;
    gap: var(--space-1);
    padding: var(--space-3) var(--space-4) var(--space-3) 3.5rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-medium);
    background: var(--color-surface);
    transition: border-color 0.2s ease;
  }

  .steps li:hover {
    border-color: var(--color-accent);
  }

  .steps li::before {
    content: counter(step);
    position: absolute;
    inset-block-start: var(--space-3);
    inset-inline-start: var(--space-4);
    display: grid;
    place-items: center;
    inline-size: 1.75rem;
    block-size: 1.75rem;
    border-radius: 50%;
    background: var(--color-accent);
    color: var(--color-surface);
    font-size: 0.875rem;
    font-weight: 700;
  }

  .steps h3 {
    font-size: 1rem;
  }

  .steps p {
    color: var(--color-text-muted);
    font-size: 0.9375rem;
  }

  .subheading {
    margin-block-start: var(--space-3);
  }

  .terminal,
  .code {
    overflow: hidden;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-medium);
    background: var(--color-surface-sunken);
  }

  .terminal-bar,
  .code-bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border-block-end: 1px solid var(--color-border);
    background: var(--color-surface);
    color: var(--color-text-muted);
    font-size: 0.8125rem;
  }

  .code-bar {
    justify-content: space-between;
  }

  .filename {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    color: var(--color-text);
  }

  .actions {
    display: flex;
    gap: var(--space-4);
  }

  .actions a,
  .actions button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0;
    border: 0;
    background: none;
    color: var(--color-accent);
    font: inherit;
    cursor: pointer;
  }

  .actions a:hover,
  .actions button:hover {
    text-decoration: underline;
  }

  pre {
    margin: 0;
    padding: var(--space-4);
    overflow-x: auto;
    font-size: 0.8125rem;
    line-height: 1.6;
  }

  /* Colors of the Python logo: blue and yellow. */
  pre {
    --python-blue: light-dark(oklch(45% 0.11 250), oklch(72% 0.11 250));
    --python-blue-soft: light-dark(oklch(55% 0.12 245), oklch(80% 0.09 245));
    --python-yellow: light-dark(oklch(52% 0.12 85), oklch(88% 0.16 95));
  }

  .keyword {
    color: var(--python-blue);
    font-weight: 600;
  }

  .builtin,
  .number {
    color: var(--python-blue-soft);
  }

  .string {
    color: var(--python-yellow);
  }

  .comment {
    color: var(--color-text-muted);
    font-style: italic;
  }

  .prompt {
    color: var(--color-accent);
    user-select: none;
  }

  .output {
    color: var(--tone-positive, var(--color-text));
    font-weight: 600;
  }
</style>
