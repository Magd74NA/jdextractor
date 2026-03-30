<script lang="ts">
  import uPlot from 'uplot';
  import type { Job } from '../lib/types';
  import { computeScoreDistribution } from '../lib/dashboard';
  import { getChartColors } from '../lib/theme';
  import UPlotChart from './UPlotChart.svelte';

  let { jobs }: { jobs: Job[] } = $props();

  let containerEl: HTMLDivElement;
  let width = $state(600);
  let colors = $state<ReturnType<typeof getChartColors> | null>(null);

  let chartData = $derived(computeScoreDistribution(jobs));

  let opts = $derived.by((): uPlot.Options | null => {
    if (!colors) return null;
    const c = colors;
    return {
      width,
      height: 200,
      scales: {
        x: {
          time: false,
          range: [0.5, 10.5],
        },
      },
      axes: [
        {
          stroke: c.muted,
          grid: { show: false },
          values: (_u: uPlot, vals: number[]) => vals.map(v => String(Math.round(v))),
          splits: (_u: uPlot) => [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
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
          label: 'Jobs',
          fill: c.primary + '80',
          stroke: c.primary,
          width: 1,
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
  <h4>Score Distribution</h4>
  {#if chartData[0].length > 0 && opts}
    <UPlotChart {opts} data={chartData} />
  {:else}
    <p class="empty">No scores yet.</p>
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
