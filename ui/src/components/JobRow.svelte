<script lang="ts">
  import { api } from '../lib/api';
  import { refreshJobs } from '../lib/stores.svelte';
  import type { Job, JobStatus } from '../lib/types';
  import { JOB_STATUSES } from '../lib/types';

  let { job }: { job: Job } = $props();

  let expanded = $state(false);
  let filesLoading = $state(false);
  let resume = $state('');
  let cover = $state<string | undefined>(undefined);
  let resumeSaved = $state(false);
  let coverSaved = $state(false);

  function scoreBadgeClass(score: number): string {
    if (score >= 7) return 'badge-good';
    if (score >= 5) return 'badge-ok';
    return 'badge-low';
  }

  async function toggle() {
    expanded = !expanded;
    if (expanded && !resume) {
      filesLoading = true;
      try {
        const files = await api.getJobFiles(job.dir);
        resume = files.resume;
        cover = files.cover;
      } finally {
        filesLoading = false;
      }
    }
  }

  async function updateStatus(e: Event) {
    const status = (e.target as HTMLSelectElement).value as JobStatus;
    await api.updateJobStatus(job.dir, status);
  }

  async function saveResume() {
    await api.saveJobFiles(job.dir, { resume });
    resumeSaved = true;
    setTimeout(() => resumeSaved = false, 3000);
  }

  async function saveCover() {
    if (cover === undefined) return;
    await api.saveJobFiles(job.dir, { cover });
    coverSaved = true;
    setTimeout(() => coverSaved = false, 3000);
  }

  async function deleteJob() {
    await api.deleteJob(job.dir);
    await refreshJobs();
  }
</script>

<tr>
  <td>{job.date}</td>
  <td>{job.company}</td>
  <td>{job.role}</td>
  <td><span class="badge {scoreBadgeClass(job.score)}">{job.score}</span></td>
  <td>
    <select value={job.status} onchange={updateStatus}>
      {#each JOB_STATUSES as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
  </td>
  <td>
    <button class="outline" onclick={toggle}>{expanded ? '▲' : '▼'}</button>
    <button class="outline danger-btn" onclick={deleteJob}>Delete</button>
  </td>
</tr>

{#if expanded}
  <tr class="expanded-row">
    <td colspan="6">
      {#if filesLoading}
        <p aria-busy="true">Loading files...</p>
      {:else}
        <label>
          Resume
          <textarea rows={6} bind:value={resume}></textarea>
        </label>
        <button onclick={saveResume}>Save Resume</button>
        {#if resumeSaved}<small class="success">Saved!</small>{/if}

        {#if cover !== undefined}
          <label>
            Cover Letter
            <textarea rows={6} bind:value={cover}></textarea>
          </label>
          <button onclick={saveCover}>Save Cover</button>
          {#if coverSaved}<small class="success">Saved!</small>{/if}
        {/if}
      {/if}
    </td>
  </tr>
{/if}

<style>
  .badge {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-weight: 600;
    font-size: 0.85rem;
  }

  .badge-good {
    background: #bbf7d0;
    color: #166534;
  }

  .badge-ok {
    background: #fef08a;
    color: #854d0e;
  }

  .badge-low {
    background: #fecaca;
    color: #991b1b;
  }

  .danger-btn {
    color: var(--pico-del-color);
    border-color: var(--pico-del-color);
  }

  .expanded-row td {
    padding: 1rem;
  }

  .success {
    color: var(--pico-ins-color);
    margin-left: 0.5rem;
  }

  select {
    margin-bottom: 0;
    padding: 0.25rem 0.5rem;
  }

  td button {
    margin-bottom: 0;
    padding: 0.25rem 0.5rem;
  }
</style>
