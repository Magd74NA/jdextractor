<script lang="ts">
  import Router from "svelte-spa-router";
  import Home from "./routes/Home.svelte";
  import HelloWorld from "./routes/HelloWorld.svelte";

  const routes = {
    "/": Home,
    "/hello": HelloWorld,
  };
</script>

<Router {routes} />
