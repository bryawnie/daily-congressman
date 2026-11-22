<script lang="ts">
  import { animate, stagger } from "motion";

  const CENTER_X = 100;
  const CENTER_Y = 88;
  const rows = [
    { radius: 26, seats: 5 },
    { radius: 44, seats: 8 },
    { radius: 62, seats: 11 },
  ];
  const HIGHLIGHT = { row: 1, seat: 3 };

  const seats = rows.flatMap(({ radius, seats }, row) =>
    Array.from({ length: seats }, (_, seat) => {
      const angle = Math.PI * (1 - seat / (seats - 1));
      return {
        x: CENTER_X + radius * Math.cos(angle),
        y: CENTER_Y - radius * Math.sin(angle),
        highlighted: row === HIGHLIGHT.row && seat === HIGHLIGHT.seat,
      };
    }),
  );
  const chosen = seats.find((seat) => seat.highlighted);

  function intro(svg: SVGSVGElement) {
    if (matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    const dots = svg.querySelectorAll(".seat");
    const halo = svg.querySelector(".halo");
    const chosenSeat = svg.querySelector(".chosen");

    animate(
      dots,
      { opacity: [0, 1], scale: [0.3, 1] },
      { duration: 0.5, delay: stagger(0.018), ease: "backOut" },
    );
    const controls: { stop: () => void }[] = [];
    if (chosenSeat) {
      controls.push(
        animate(chosenSeat, { scale: [0, 1.25, 1] }, { duration: 0.7, delay: 0.7 }),
      );
    }
    if (halo) {
      controls.push(
        animate(
          halo,
          { scale: [1, 1.6, 1], opacity: [0.9, 0.2, 0.9] },
          { duration: 2.6, delay: 1.4, repeat: Infinity, ease: "easeInOut" },
        ),
      );
    }
    return { destroy: () => controls.forEach((control) => control.stop()) };
  }
</script>

<svg viewBox="30 18 140 76" aria-hidden="true" focusable="false" use:intro>
  {#each seats as seat}
    {#if !seat.highlighted}
      <circle class="seat" cx={seat.x} cy={seat.y} r="3.8" />
    {/if}
  {/each}
  {#if chosen}
    <circle class="halo" cx={chosen.x} cy={chosen.y} r="10" />
    <circle class="chosen" cx={chosen.x} cy={chosen.y} r="6" />
  {/if}
</svg>

<style>
  svg {
    inline-size: 4.5rem;
    block-size: auto;
    flex: none;
  }

  circle {
    transform-box: fill-box;
    transform-origin: center;
  }

  .seat {
    fill: color-mix(in oklch, var(--color-text) 38%, transparent);
  }

  .chosen {
    fill: var(--color-highlight);
  }

  .halo {
    fill: color-mix(in oklch, var(--color-highlight) 28%, transparent);
  }
</style>
