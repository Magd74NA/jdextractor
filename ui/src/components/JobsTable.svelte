<script lang="ts">
  import { getJobs, loadJobs } from '../lib/stores.svelte';
  import JobRow from './JobRow.svelte';

  let loading = $state(true);
  let error = $state('');
  let jobs = $derived(getJobs());

  async function init() {
    try {
      await loadJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load jobs';
    } finally {
      loading = false;
    }
  }

  init();
</script>

<section class="jobs-section">
  <h3>Applications</h3>

  {#if loading}
    <p aria-busy="true">Loading applications...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if jobs.length === 0}
    <p>No applications yet.</p>
  {:else}
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Date</th>
            <th>Company</th>
            <th>Role</th>
            <th>Score</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each jobs as job (job.dir)}
            <JobRow {job} />
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>

<style>
  .jobs-section {
    margin-top: 1rem;
    border: 1px solid var(--pico-muted-border-color);
    border-radius: var(--pico-border-radius);
    padding: 1rem;
  }

  .jobs-section h3 {
    margin-top: 0;
  }

  .table-wrap {
    overflow-x: auto;
  }

  .error {
    color: var(--pico-del-color);
  }
</style>
