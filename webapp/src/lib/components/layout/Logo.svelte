<script lang="ts">
  // Hemicycle: three rows of seats, one of them highlighted as the chosen one.
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
</script>

<svg viewBox="30 18 140 76" aria-hidden="true" focusable="false">
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
