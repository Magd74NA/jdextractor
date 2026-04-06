<script lang="ts">
  import type { DashboardStats } from "../lib/dashboard";

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
