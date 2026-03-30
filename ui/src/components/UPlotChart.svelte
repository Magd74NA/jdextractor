<script lang="ts">
  import uPlot from 'uplot';
  import 'uplot/dist/uPlot.min.css';

  let { opts, data }: { opts: uPlot.Options; data: uPlot.AlignedData } = $props();

  let container: HTMLDivElement;
  let chart: uPlot | undefined;

  $effect(() => {
    if (!container || data[0].length === 0) return;

    chart?.destroy();
    chart = new uPlot(opts, data, container);

    return () => {
      chart?.destroy();
      chart = undefined;
    };
  });
</script>

<div bind:this={container} class="uplot-wrap"></div>

<style>
  .uplot-wrap {
    width: 100%;
  }

  /* Neutralize Pico CSS table styles inside uPlot legend */
  .uplot-wrap :global(.u-legend) {
    text-align: left;
  }

  .uplot-wrap :global(.u-legend th),
  .uplot-wrap :global(.u-legend td) {
    padding: 0 0.5rem 0 0;
    border: none;
    font-size: 0.8rem;
    white-space: nowrap;
  }

  .uplot-wrap :global(.u-legend table) {
    margin: 0;
    table-layout: auto;
  }
</style>
