<script lang="ts">
  import type { NetworkingStats } from "../lib/dashboard";

  let { stats }: { stats: NetworkingStats } = $props();
</script>

<div class="stat-grid">
  <article class="stat-card">
    <h3>{stats.totalContacts}</h3>
    <small>Contacts</small>
  </article>
  <article class="stat-card">
    <h3>{stats.activeContacts}</h3>
    <small>Active</small>
  </article>
  <article
    class="stat-card overdue-card"
    class:has-overdue={stats.overdueFollowups > 0}
  >
    <h3>{stats.overdueFollowups}</h3>
    <small>Overdue</small>
  </article>
  <article class="stat-card">
    <h3>{stats.upcomingFollowups}</h3>
    <small>Due This Week</small>
  </article>
</div>

<style>
  /* Component-specific styles only - shared styles moved to app.css */
  .overdue-card.has-overdue h3 {
    color: var(--pico-del-color);
  }
</style>
