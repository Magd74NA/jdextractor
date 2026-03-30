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

<style>
  .table-wrap {
    overflow-x: auto;
  }

  .error {
    color: var(--pico-del-color);
  }
</style>
