<script lang="ts">
  import ContactsTable from '../components/ContactsTable.svelte';
  import { api } from '../lib/api';
  import { refreshContacts } from '../lib/stores.svelte';
  import { CONTACT_STATUSES, CHANNELS } from '../lib/types';

  let showForm = $state(false);
  let saving = $state(false);
  let formError = $state('');

  let name = $state('');
  let company = $state('');
  let role = $state('');
  let email = $state('');
  let phone = $state('');
  let linkedin = $state('');
  let source = $state('');
  let status = $state('new');
  let followUpDate = $state('');
  let notes = $state('');

  function resetForm() {
    name = ''; company = ''; role = ''; email = ''; phone = '';
    linkedin = ''; source = ''; status = 'new'; followUpDate = ''; notes = '';
    formError = '';
  }

  async function createContact() {
    if (!name.trim()) { formError = 'Name is required'; return; }
    saving = true;
    formError = '';
    try {
      const data: Record<string, string> = { name: name.trim(), status };
      if (company) data.company = company;
      if (role) data.role = role;
      if (email) data.email = email;
      if (phone) data.phone = phone;
      if (linkedin) data.linkedin = linkedin;
      if (source) data.source = source;
      if (followUpDate) data.follow_up_date = followUpDate;
      if (notes) data.notes = notes;
      await api.createContact(data);
      resetForm();
      showForm = false;
      await refreshContacts();
    } catch (e) {
      formError = e instanceof Error ? e.message : 'Failed to create contact';
    } finally {
      saving = false;
    }
  }
</script>

<div class="page-header">
  <h2>Contacts</h2>
  <button class="outline btn-sm" onclick={() => { showForm = !showForm; if (!showForm) resetForm(); }}>
    {showForm ? 'Cancel' : '+ New Contact'}
  </button>
</div>

{#if showForm}
  <article class="contact-form">
    <h3>New Contact</h3>
    {#if formError}<p class="error">{formError}</p>{/if}
    <div class="form-grid">
      <label>Name * <input bind:value={name} placeholder="Jane Doe" /></label>
      <label>Company <input bind:value={company} placeholder="Acme Corp" /></label>
      <label>Role <input bind:value={role} placeholder="Engineering Manager" /></label>
      <label>Email <input type="email" bind:value={email} placeholder="jane@acme.com" /></label>
      <label>Phone <input bind:value={phone} placeholder="+1 555 0100" /></label>
      <label>LinkedIn <input bind:value={linkedin} placeholder="https://linkedin.com/in/..." /></label>
      <label>How we met
        <input bind:value={source} placeholder="conference, referral, cold outreach…" />
      </label>
      <label>Status
        <select bind:value={status}>
          {#each CONTACT_STATUSES as s}
            <option value={s}>{s}</option>
          {/each}
        </select>
      </label>
      <label>Follow-up date <input type="date" bind:value={followUpDate} /></label>
      <label class="full-width">Notes <textarea rows={2} bind:value={notes}></textarea></label>
    </div>
    <button onclick={createContact} aria-busy={saving} disabled={saving}>Create Contact</button>
  </article>
{/if}

<ContactsTable />

<style>
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .page-header h2 {
    margin: 0;
  }

  .btn-sm {
    padding: 0.3em 0.75em;
    margin-bottom: 0;
  }

  .contact-form {
    margin-bottom: 2rem;
    padding: 1.5rem;
  }

  .contact-form h3 {
    margin-top: 0;
    margin-bottom: 1rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .full-width {
    grid-column: span 2;
  }

  .form-grid label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.85rem;
    font-weight: 600;
    margin-bottom: 0;
  }

  .form-grid input,
  .form-grid select,
  .form-grid textarea {
    margin-bottom: 0;
    font-size: 0.85rem;
    font-weight: 400;
  }

  .error {
    color: var(--pico-del-color);
    font-size: 0.85rem;
  }
</style>
