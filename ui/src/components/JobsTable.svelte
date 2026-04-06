<script lang="ts">
  import { getJobs, loadJobs, loadContacts } from "../lib/stores.svelte";
  import JobRow from "./JobRow.svelte";
  import type { Job } from "../lib/types";

  let { jobs: jobsProp }: { jobs?: Job[] } = $props();

  let loading = $state(true);
  let error = $state("");
  let allJobs = $derived(getJobs());
  let jobs = $derived(jobsProp ?? allJobs);

  async function init() {
    try {
      await Promise.all([loadJobs(), loadContacts()]);
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load jobs";
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
          <th class="col-date">Date</th>
          <th class="col-company">Company</th>
          <th class="col-role">Role</th>
          <th class="col-score">Score</th>
          <th class="col-status">Status</th>
          <th class="col-actions">Actions</th>
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
  /* Component-specific styles only - shared styles moved to app.css */
  .col-date {
    width: 8em;
  }

  .col-company {
    width: 19%;
  }

  .col-role {
    width: 26%;
  }

  .col-score {
    width: 3.5em;
    text-align: center;
  }

  .col-status {
    width: 12em;
  }

  .col-actions {
    width: 9.5em;
  }
</style>
