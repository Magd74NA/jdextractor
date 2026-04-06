<script lang="ts">
  import { getContacts, loadContacts } from "../lib/stores.svelte";
  import ContactRow from "./ContactRow.svelte";

  let loading = $state(true);
  let error = $state("");
  let contacts = $derived(getContacts());

  async function init() {
    try {
      await loadContacts();
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load contacts";
    } finally {
      loading = false;
    }
  }

  init();
</script>

{#if loading}
  <p aria-busy="true">Loading contacts...</p>
{:else if error}
  <p class="error">{error}</p>
{:else if contacts.length === 0}
  <p>No contacts yet.</p>
{:else}
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th class="col-name">Name</th>
          <th class="col-company">Company</th>
          <th class="col-status">Status</th>
          <th class="col-followup">Follow-up</th>
          <th class="col-convos">Convos</th>
          <th class="col-actions">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each contacts as contact (contact.dir)}
          <ContactRow {contact} />
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  /* Component-specific styles only - shared styles moved to app.css */
  .col-name {
    width: 18%;
  }

  .col-company {
    width: 20%;
  }

  .col-status {
    width: 14em;
  }

  .col-followup {
    width: 9em;
  }

  .col-convos {
    width: 4em;
    text-align: center;
  }

  .col-actions {
    width: 9.5em;
  }
</style>
