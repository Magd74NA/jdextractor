<script lang="ts">
  import { getJobs, loadJobs } from '../lib/stores.svelte';
  import { computeStats, computeStatusCounts } from '../lib/dashboard';
  import StatCards from '../components/StatCards.svelte';
  import ActivityChart from '../components/ActivityChart.svelte';
  import ScoreChart from '../components/ScoreChart.svelte';
  import StatusBar from '../components/StatusBar.svelte';

  let loading = $state(true);
  let error = $state('');
  let jobs = $derived(getJobs());
  let stats = $derived(computeStats(jobs));
  let statusCounts = $derived(computeStatusCounts(jobs));

  async function init() {
    try {
      await loadJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load';
    } finally {
      loading = false;
    }
  }

  init();
</script>

<div class="page-header">
  <h2>Dashboard</h2>
</div>

{#if loading}
  <p aria-busy="true">Loading dashboard...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <StatCards {stats} />

  <section class="chart-section">
    <ActivityChart {jobs} />
  </section>

  <section class="chart-section">
    <StatusBar counts={statusCounts} />
  </section>

  <section class="chart-section">
    <ScoreChart {jobs} />
  </section>
{/if}

<style>
  .chart-section {
    margin-bottom: 2rem;
  }

  .error {
    color: var(--pico-del-color);
  }
</style>
