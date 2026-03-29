<script lang="ts">
  import Router from "svelte-spa-router";
  import HelloWorld from "./routes/HelloWorld.svelte";

  const routes = {
    "/": HelloWorld,
  };
</script>

<Router {routes} />
