<script lang="ts">
  import uPlot from 'uplot';
  import type { Job } from '../lib/types';
  import { computeActivityData } from '../lib/dashboard';
  import { getChartColors } from '../lib/theme';
  import UPlotChart from './UPlotChart.svelte';

  let { jobs }: { jobs: Job[] } = $props();

  let containerEl: HTMLDivElement;
  let width = $state(600);
  let colors = $state<ReturnType<typeof getChartColors> | null>(null);

  const DAY_S = 86400;
  const MIN_WINDOW = 14 * DAY_S;

  let chartData = $derived(computeActivityData(jobs));

  let xRange = $derived.by((): [number, number] => {
    const xs = chartData[0];
    if (xs.length === 0) return [0, MIN_WINDOW];
    const lo = xs[0]!;
    const hi = xs[xs.length - 1]!;
    const span = Math.max(hi - lo, MIN_WINDOW);
    return [lo - DAY_S * 0.5, lo + span + DAY_S * 1.5];
  });

  let opts = $derived.by((): uPlot.Options | null => {
    if (!colors) return null;
    const c = colors;
    return {
      width,
      height: 250,
      scales: {
        x: { time: true, range: xRange },
      },
      axes: [
        {
          stroke: c.muted,
          grid: { stroke: c.border, width: 1 },
          ticks: { stroke: c.border },
        },
        {
          stroke: c.muted,
          grid: { stroke: c.border, width: 1 },
          ticks: { stroke: c.border },
        },
      ],
      series: [
        {},
        {
          label: 'Daily',
          fill: c.primary + '40',
          stroke: c.primary,
          width: 2,
          paths: uPlot.paths.bars!({ size: [0.6, 100] }),
        },
      ],
    };
  });

  $effect(() => {
    colors = getChartColors();
  });

  $effect(() => {
    if (!containerEl) return;
    const ro = new ResizeObserver(entries => {
      const entry = entries[0];
      if (entry) width = entry.contentRect.width;
    });
    ro.observe(containerEl);
    return () => ro.disconnect();
  });
</script>

<div bind:this={containerEl} class="chart-wrapper">
  <h4>Applications Over Time</h4>
  {#if chartData[0].length > 0 && opts}
    <UPlotChart {opts} data={chartData} />
  {:else}
    <p class="empty">No application data yet.</p>
  {/if}
</div>

<style>
  h4 {
    margin-bottom: 0.75rem;
  }

  .empty {
    color: var(--pico-muted-color);
  }
</style>
