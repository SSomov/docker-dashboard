<script>
  import ContainerCard from "./ContainerCard.svelte";

  export let filteredGroups = [];
  export let containerStats = new Map();
  export let logsShow = false;
  export let containerRestart = false;
  export let onOpenLogs = (containerId, containerName) => {};
  export let onRestartContainer = (containerId, containerName) => {};
  export let loading = false;
  export let loadingStatus = "Загрузка...";
  export let totalFixedHeight = 0;
  export let viewMode = "tile";

  // Получение короткого ID контейнера
  const getShortId = (id) => {
    return id ? id.substring(0, 12) : "";
  };

  // Парсинг CPU из строки типа "2.0"
  const parseCPU = (cpuString) => {
    if (!cpuString) return 0;
    const match = cpuString.match(/^([\d.]+)/);
    return match ? parseFloat(match[1]) : 0;
  };

  // Парсинг Memory из строки типа "4G", "2.0G", "512M"
  const parseMemory = (memoryString) => {
    if (!memoryString) return 0;
    const match = memoryString.match(/^([\d.]+)([GMK]?)$/i);
    if (!match) return 0;
    const value = parseFloat(match[1]);
    const unit = (match[2] || "").toUpperCase();
    switch (unit) {
      case "G":
        return value * 1024; // GB to MB
      case "M":
        return value;
      case "K":
        return value / 1024; // KB to MB
      default:
        return value / (1024 * 1024); // bytes to MB
    }
  };

  // Форматирование Memory обратно в строку
  const formatMemory = (mb) => {
    if (mb >= 1024) {
      return (mb / 1024).toFixed(1) + " GB";
    }
    return mb.toFixed(1) + " MB";
  };

  // Форматирование CPU для отображения
  const formatCPU = (container, stats) => {
    const hasCpuLimit =
      container.DeployResources &&
      (container.DeployResources.CPULimit ||
        container.DeployResources.CPUReservation);
    const hasCpuStats = stats && typeof stats.cpu === "number";

    if (hasCpuLimit) {
      const cpuLimit = parseCPU(
        container.DeployResources.CPULimit || "0",
      );
      const cpuReservation = parseCPU(
        container.DeployResources.CPUReservation || "0",
      );
      const cpuUsage = stats ? (stats.cpu || 0) : 0;
      const cpuMax = Math.max(
        cpuLimit || cpuReservation,
        cpuReservation || cpuLimit,
        1,
      );
      const usagePercent =
        cpuMax > 0 && cpuUsage >= 0 ? (cpuUsage / cpuMax) * 100 : 0;
      return `${cpuUsage.toFixed(2)} / ${cpuReservation.toFixed(1)} / ${cpuLimit.toFixed(1)}${usagePercent > 0 ? ` (${usagePercent.toFixed(1)}%)` : ""}`;
    } else if (hasCpuStats) {
      const cpuUsage = stats ? (stats.cpu || 0) : 0;
      return `${cpuUsage.toFixed(2)}`;
    }
    return "-";
  };

  // Форматирование RAM для отображения
  const formatRAM = (container, stats) => {
    const hasMemLimit =
      container.DeployResources &&
      (container.DeployResources.MemoryLimit ||
        container.DeployResources.MemoryReservation);
    const hasMemStats = stats && typeof stats.memory === "number";

    if (hasMemLimit) {
      const memLimitMB = parseMemory(
        container.DeployResources.MemoryLimit || "0",
      );
      const memReservationMB = parseMemory(
        container.DeployResources.MemoryReservation || "0",
      );
      const memUsageBytes = stats ? (stats.memory || 0) : 0;
      const memUsageMB = memUsageBytes / (1024 * 1024);
      const memMax = Math.max(
        memLimitMB || memReservationMB,
        memReservationMB || memLimitMB,
        1,
      );
      const usagePercent =
        memMax > 0 && memUsageMB >= 0 ? (memUsageMB / memMax) * 100 : 0;
      return `${formatMemory(memUsageMB)} / ${formatMemory(memReservationMB)} / ${formatMemory(memLimitMB)}${usagePercent > 0 ? ` (${usagePercent.toFixed(1)}%)` : ""}`;
    } else if (hasMemStats) {
      const memUsageBytes = stats ? (stats.memory || 0) : 0;
      const memUsageMB = memUsageBytes / (1024 * 1024);
      return formatMemory(memUsageMB);
    }
    return "-";
  };
</script>

<div class="content-wrapper" style="padding-top: {totalFixedHeight}px">
  {#if loading}
    <div class="container">
      <div class="loader-container">
        <div class="loader"></div>
        <p class="loading-status">{loadingStatus}</p>
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
        {#if viewMode === "tile"}
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
        {:else}
          <div class="table-container">
            <table class="containers-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>ID</th>
                  <th>Image</th>
                  <th>Status</th>
                  <th>Health</th>
                  <th>Uptime</th>
                  <th>CPU</th>
                  <th>RAM</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredGroups as group (group.project_name || "ungrouped")}
                  {#if group.project_name}
                    <tr class="group-header-row">
                      <td colspan="9" class="group-header-cell">
                        Project: {group.project_name}
                      </td>
                    </tr>
                  {/if}
                  {#each group.containers as container (container.ID)}
                    {@const stats = containerStats.get(container.ID)}
                    <tr class="table-row" class:unhealthy={container.Health === "unhealthy"} class:stopped={!container.Run || container.State !== "running"}>
                      <td class="name-cell">
                        <strong>{container.Name}</strong>
                      </td>
                      <td class="id-cell">{getShortId(container.ID)}</td>
                      <td class="image-cell">{container.Image}</td>
                      <td class="status-cell">
                        <span
                          class="status-badge"
                          class:status-running={container.Run && container.State === "running" && container.Health !== "unhealthy"}
                          class:status-stopped={!container.Run || container.State !== "running"}
                          class:status-unhealthy={container.Health === "unhealthy"}
                        >
                          {container.State}
                        </span>
                      </td>
                      <td class="health-cell">
                        {#if container.Health}
                          <span class="health-badge" class:unhealthy={container.Health === "unhealthy"}>
                            {container.Health}
                          </span>
                        {:else}
                          <span class="health-badge no-health">-</span>
                        {/if}
                      </td>
                      <td class="uptime-cell">{container.Uptime || "-"}</td>
                      <td class="cpu-cell">{formatCPU(container, stats)}</td>
                      <td class="ram-cell">{formatRAM(container, stats)}</td>
                      <td class="actions-cell">
                        <div class="table-actions">
                          {#if logsShow}
                            <button
                              class="action-button logs-button"
                              on:click={() => onOpenLogs(container.ID, container.Name)}
                              title="View logs"
                            >
                              logs
                            </button>
                          {/if}
                          {#if containerRestart}
                            <button
                              class="action-button restart-button"
                              on:click={() => onRestartContainer(container.ID, container.Name)}
                              title="Restart container"
                            >
                              restart
                            </button>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
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
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100vh;
    gap: 1rem;
  }

  .loading-status {
    color: rgba(255, 255, 255, 0.8);
    font-size: 1rem;
    margin-top: 1rem;
    text-align: center;
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

  .table-container {
    overflow-x: auto;
    margin: 2rem;
    background-color: rgba(76, 175, 80, 0.05);
    border: 2px solid #4caf50;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }

  .containers-table {
    width: 100%;
    border-collapse: collapse;
    background-color: white;
    border-radius: 4px;
    overflow: hidden;
  }

  .containers-table thead {
    background-color: #4caf50;
    color: white;
  }

  .containers-table th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
  }

  .containers-table tbody tr {
    border-bottom: 1px solid #e0e0e0;
    transition: background-color 0.2s;
  }

  .containers-table tbody tr:hover {
    background-color: #f5f5f5;
  }

  .containers-table tbody tr.unhealthy {
    background-color: #ffebee;
  }

  .containers-table tbody tr.stopped {
    background-color: #f5f5f5;
    opacity: 0.7;
  }

  .containers-table tbody tr.stopped td {
    color: #9e9e9e;
  }

  .group-header-row {
    background-color: #4caf50;
  }

  .group-header-row:hover {
    background-color: #4caf50;
  }

  .group-header-cell {
    background-color: #4caf50;
    color: white;
    font-weight: 600;
    font-size: 1.1rem;
    padding: 0.75rem 1rem;
  }

  .containers-table td {
    padding: 0.75rem 1rem;
    font-size: 0.9rem;
    vertical-align: middle;
  }

  .name-cell {
    font-weight: 500;
    min-width: 150px;
  }

  .id-cell {
    font-family: monospace;
    font-size: 0.85rem;
    color: #666;
    min-width: 100px;
  }

  .image-cell {
    min-width: 200px;
    word-break: break-word;
  }

  .status-cell {
    min-width: 100px;
  }

  .health-cell {
    min-width: 100px;
  }

  .uptime-cell {
    min-width: 120px;
  }

  .cpu-cell {
    min-width: 180px;
    font-size: 0.85rem;
    font-family: monospace;
  }

  .ram-cell {
    min-width: 180px;
    font-size: 0.85rem;
    font-family: monospace;
  }

  .actions-cell {
    min-width: 150px;
  }

  .status-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-badge.status-running {
    background-color: #4caf50;
    color: white;
  }

  .status-badge.status-stopped {
    background-color: #9e9e9e;
    color: white;
  }

  .status-badge.status-unhealthy {
    background-color: #f44336;
    color: white;
  }

  .health-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: capitalize;
    background-color: #4caf50;
    color: white;
  }

  .health-badge.unhealthy {
    background-color: #f44336;
  }

  .health-badge.no-health {
    background-color: #e0e0e0;
    color: #666;
  }

  .table-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .action-button {
    padding: 0.4rem 0.8rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    font-weight: 600;
    transition: background-color 0.2s;
    white-space: nowrap;
  }

  .action-button.logs-button {
    background-color: #4caf50;
    color: white;
  }

  .action-button.logs-button:hover {
    background-color: #45a049;
  }

  .action-button.restart-button {
    background-color: #2196f3;
    color: white;
  }

  .action-button.restart-button:hover {
    background-color: #1976d2;
  }
</style>
