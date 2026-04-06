<script lang="ts">
  import { api } from "../lib/api";
  import {
    refreshJobs,
    getContacts,
    refreshContacts,
  } from "../lib/stores.svelte";
  import type { Job, JobStatus, Contact } from "../lib/types";
  import { JOB_STATUSES } from "../lib/types";

  let linkedContacts = $derived(
    getContacts().filter((c) => (c.linked_jobs ?? []).includes(job.dir)),
  );

  let { job }: { job: Job } = $props();

  let expanded = $state(false);
  let filesLoading = $state(false);
  let resume = $state("");
  let cover = $state<string | undefined>(undefined);
  let resumeSaved = $state(false);
  let coverSaved = $state(false);

  let editing = $state(false);
  let editDate = $state("");
  let editCompany = $state("");
  let editRole = $state("");

  function startEdit() {
    editDate = job.date;
    editCompany = job.company;
    editRole = job.role;
    editing = true;
  }

  function cancelEdit() {
    editing = false;
  }

  async function saveMeta() {
    await api.updateJobMeta(job.dir, {
      company: editCompany,
      role: editRole,
      date: editDate,
    });
    editing = false;
    await refreshJobs();
  }

  function scoreBadgeClass(score: number): string {
    if (score >= 7) return "badge-good";
    if (score >= 5) return "badge-ok";
    return "badge-low";
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
    setTimeout(() => (resumeSaved = false), 3000);
  }

  async function saveCover() {
    if (cover === undefined) return;
    await api.saveJobFiles(job.dir, { cover });
    coverSaved = true;
    setTimeout(() => (coverSaved = false), 3000);
  }

  async function deleteJob() {
    await api.deleteJob(job.dir);
    await refreshJobs();
  }

  async function unlinkContact(contact: Contact) {
    const updated = (contact.linked_jobs ?? []).filter((j) => j !== job.dir);
    await api.updateContact(contact.dir, { linked_jobs: updated });
    await refreshContacts();
  }
</script>

<tr>
  <td class="truncate">
    {#if editing}<input
        class="edit-input"
        bind:value={editDate}
      />{:else}{job.date}{/if}
  </td>
  <td class="truncate">
    {#if editing}<input
        class="edit-input"
        bind:value={editCompany}
      />{:else}{job.company}{/if}
  </td>
  <td class="truncate">
    {#if editing}<input
        class="edit-input"
        bind:value={editRole}
      />{:else}{job.role}{/if}
  </td>
  <td class="score-cell">
    <span class="badge {scoreBadgeClass(job.score)}">{job.score}</span>
  </td>
  <td>
    <select value={job.status} onchange={updateStatus}>
      {#each JOB_STATUSES as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
  </td>
  <td class="actions-cell">
    <button class="outline btn-sm" onclick={toggle}
      >{expanded ? "▲" : "▼"}</button
    >
    {#if editing}
      <button class="outline btn-sm save-btn" onclick={saveMeta} title="Save"
        >💾</button
      >
      <button class="outline btn-sm" onclick={cancelEdit} title="Cancel"
        >✕</button
      >
    {:else}
      <button class="outline btn-sm" onclick={startEdit} title="Edit">✏</button
      >
      <button
        class="outline btn-sm danger-btn"
        onclick={deleteJob}
        title="Delete">✕</button
      >
    {/if}
  </td>
</tr>

{#if expanded}
  <tr class="expanded-row">
    <td colspan="6">
      {#if filesLoading}
        <p aria-busy="true">Loading files...</p>
      {:else}
        <div class="file-section">
          <div class="file-header">
            <h4>Resume</h4>
            <div class="file-actions">
              {#if resumeSaved}<small class="success">Saved!</small>{/if}
              <button class="outline btn-sm" onclick={saveResume}>Save</button>
            </div>
          </div>
          <textarea class="mono" rows={8} bind:value={resume}></textarea>
        </div>

        {#if cover !== undefined}
          <div class="file-section">
            <div class="file-header">
              <h4>Cover Letter</h4>
              <div class="file-actions">
                {#if coverSaved}<small class="success">Saved!</small>{/if}
                <button class="outline btn-sm" onclick={saveCover}>Save</button>
              </div>
            </div>
            <textarea class="mono" rows={8} bind:value={cover}></textarea>
          </div>
        {/if}

        {#if linkedContacts.length > 0}
          <div class="file-section">
            <div class="file-header">
              <h4>Linked Contacts</h4>
            </div>
            <div class="linked-tags">
              {#each linkedContacts as c}
                <span class="tag contact-tag">
                  <a href="#/contacts">{c.name}</a>
                  {#if c.company}<span class="contact-company"
                      >@ {c.company}</span
                    >{/if}
                  <button
                    class="tag-remove"
                    onclick={() => unlinkContact(c)}
                    title="Unlink">x</button
                  >
                </span>
              {/each}
            </div>
          </div>
        {/if}
      {/if}
    </td>
  </tr>
{/if}

<style>
  /* Component-specific styles only - shared styles moved to app.css */
  .score-cell {
    text-align: center;
    overflow: visible;
  }

  .linked-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .contact-tag {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }

  .contact-tag a {
    text-decoration: none;
    font-weight: 600;
  }

  .contact-company {
    color: var(--pico-muted-color);
    font-size: 0.72rem;
  }
</style>
