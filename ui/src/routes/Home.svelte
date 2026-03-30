<script lang="ts">
  import { loadConfig, loadPromptConfig } from '../lib/stores.svelte';
  import ConfigCard from '../components/ConfigCard.svelte';
  import TemplatesCard from '../components/TemplatesCard.svelte';
  import ProcessUrl from '../components/ProcessUrl.svelte';
  import ProcessBatch from '../components/ProcessBatch.svelte';
  import ProcessLocal from '../components/ProcessLocal.svelte';
  import JobsTable from '../components/JobsTable.svelte';

  async function init() {
    await Promise.all([loadConfig(), loadPromptConfig()]);
  }

  init();
</script>

<h1>jdextractor</h1>

<ConfigCard />
<TemplatesCard />
<ProcessUrl />
<ProcessBatch />
<ProcessLocal />
<JobsTable />
