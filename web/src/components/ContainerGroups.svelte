<script>
  import ContainerCard from "./ContainerCard.svelte";

  export let filteredGroups = [];
  export let containerStats = new Map();
  export let logsShow = false;
  export let containerRestart = false;
  export let onOpenLogs = (containerId, containerName) => {};
  export let onRestartContainer = (containerId, containerName) => {};
  export let loading = false;
  export let totalFixedHeight = 0;
</script>

<div class="content-wrapper" style="padding-top: {totalFixedHeight}px">
  {#if loading}
    <div class="container">
      <div class="loader-container">
        <div class="loader"></div>
      </div>
    </div>
  {:else}
    <div class="container">
      {#if filteredGroups.length === 0}
        <div class="empty-message">
          <p>No containers found</p>
          <p class="empty-hint">
            Check your filters or wait for containers to load
          </p>
        </div>
      {:else}
        {#each filteredGroups as group (group.project_name || "ungrouped")}
          <div class="container-group">
            {#if group.project_name}
              <div class="group-header">
                Project: {group.project_name}
              </div>
            {/if}
            <div class="group-containers">
              {#each group.containers as container (container.ID)}
                <ContainerCard
                  {container}
                  stats={containerStats.get(container.ID)}
                  {logsShow}
                  {containerRestart}
                  {onOpenLogs}
                  {onRestartContainer}
                />
              {/each}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style>
  .content-wrapper {
    min-height: 100vh;
  }

  .container {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    margin: 2rem;
  }

  .loader {
    border: 16px solid #f3f3f3;
    border-radius: 50%;
    border-top: 16px solid #3498db;
    width: 120px;
    height: 120px;
    animation: spin 2s linear infinite;
    margin: auto;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .loader-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
  }

  .container-group {
    border: 2px solid #4caf50;
    border-radius: 8px;
    margin: 1.2rem;
    padding: 1rem;
    background-color: rgba(76, 175, 80, 0.05);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }

  .group-header {
    background-color: #4caf50;
    color: white;
    padding: 0.75rem 1rem;
    margin: -1rem -1rem 1rem -1rem;
    border-radius: 6px 6px 0 0;
    font-weight: 600;
    font-size: 1.1rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .group-containers {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .empty-message {
    text-align: center;
    padding: 4rem 2rem;
    color: rgba(255, 255, 255, 0.7);
  }

  .empty-message p {
    margin: 0.5rem 0;
    font-size: 1.2rem;
  }

  .empty-hint {
    font-size: 0.9rem;
    color: rgba(255, 255, 255, 0.5);
  }
</style>
