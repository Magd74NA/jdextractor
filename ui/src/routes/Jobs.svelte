<script lang="ts">
  import { push } from 'svelte-spa-router';
  import JobsTable from '../components/JobsTable.svelte';
</script>

<div class="page-header">
  <h2>Applications</h2>
  <button onclick={() => push('/process')}>+ Process New</button>
</div>

<JobsTable />
