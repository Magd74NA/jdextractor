<script lang="ts">
  import { getJobs, loadJobs, getContacts, loadContacts } from '../lib/stores.svelte';
  import { computeStats, computeStatusCounts, computeNetworkingStats } from '../lib/dashboard';
  import StatCards from '../components/StatCards.svelte';
  import ActivityChart from '../components/ActivityChart.svelte';
  import ScoreChart from '../components/ScoreChart.svelte';
  import StatusBar from '../components/StatusBar.svelte';
  import NetworkingStats from '../components/NetworkingStats.svelte';
  import FollowupQueue from '../components/FollowupQueue.svelte';

  let loading = $state(true);
  let error = $state('');
  let jobs = $derived(getJobs());
  let contacts = $derived(getContacts());
  let stats = $derived(computeStats(jobs));
  let statusCounts = $derived(computeStatusCounts(jobs));
  let networkingStats = $derived(computeNetworkingStats(contacts));

  async function init() {
    try {
      await Promise.all([loadJobs(), loadContacts()]);
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

  {#if contacts.length > 0}
    <hr />
    <h3 class="section-title">Networking</h3>
    <NetworkingStats stats={networkingStats} />
    <FollowupQueue />
  {/if}
{/if}

<style>
  .chart-section {
    margin-bottom: 2rem;
  }

  .error {
    color: var(--pico-del-color);
  }

  .section-title {
    margin-bottom: 1rem;
    color: var(--pico-muted-color);
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
</style>
