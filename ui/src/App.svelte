<script lang="ts">
  import Router from "svelte-spa-router";
  import Nav from "./components/Nav.svelte";
  import Jobs from "./routes/Jobs.svelte";
  import Process from "./routes/Process.svelte";
  import Settings from "./routes/Settings.svelte";
  import { loadConfig, loadPromptConfig } from "./lib/stores.svelte";

  const routes = {
    "/": Jobs,
    "/process": Process,
    "/settings": Settings,
  };

  async function init() {
    await Promise.all([loadConfig(), loadPromptConfig()]);
  }

  init();
</script>

<Nav />
<main>
  <Router {routes} />
</main>
