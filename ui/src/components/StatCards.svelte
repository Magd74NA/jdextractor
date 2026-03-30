<script lang="ts">
  import type { DashboardStats } from '../lib/dashboard';

  let { stats }: { stats: DashboardStats } = $props();
</script>

<div class="stat-grid">
  <article class="stat-card">
    <h3>{stats.total}</h3>
    <small>Total Applications</small>
  </article>
  <article class="stat-card">
    <h3>{stats.applied}</h3>
    <small>Applied</small>
  </article>
  <article class="stat-card">
    <h3>{stats.avgScore}</h3>
    <small>Avg Score</small>
  </article>
  <article class="stat-card">
    <h3>{stats.thisWeek}</h3>
    <small>This Week</small>
  </article>
</div>

<style>
  .stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    text-align: center;
    padding: 1rem;
    margin-bottom: 0;
  }

  .stat-card h3 {
    margin-bottom: 0.25rem;
    color: var(--pico-primary);
    font-size: 2rem;
  }

  .stat-card small {
    color: var(--pico-muted-color);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.75rem;
  }
</style>
