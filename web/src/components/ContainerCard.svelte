<script>
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

export let container;
export let stats = null;
export let logsShow = false;
export let containerRestart = false;
export let onOpenLogs = (containerId, containerName) => {};
export let onRestartContainer = (containerId, containerName) => {};

$: hasCpuLimit =
	container.DeployResources &&
	(container.DeployResources.CPULimit ||
		container.DeployResources.CPUReservation);
$: hasMemLimit =
	container.DeployResources &&
	(container.DeployResources.MemoryLimit ||
		container.DeployResources.MemoryReservation);
$: hasCpuStats = stats && typeof stats.cpu !== "undefined" && stats.cpu !== null;
$: hasMemStats = stats && typeof stats.memory !== "undefined" && stats.memory !== null;
$: showResources = hasCpuLimit || hasMemLimit || hasCpuStats || hasMemStats;
</script>

<div
    class="card"
    class:unhealthy={container.Health === "unhealthy"}
    class:stopped={!container.Run || container.State !== "running"}
>
    <div class="card-header">
        <h2>{container.Name}</h2>
        <div class="card-header-buttons">
            {#if logsShow}
                <button
                    class="logs-button"
                    on:click={() => onOpenLogs(container.ID, container.Name)}
                >
                    logs
                </button>
            {/if}
            {#if containerRestart}
                <button
                    class="restart-button"
                    on:click={() =>
                        onRestartContainer(container.ID, container.Name)}
                >
                    restart
                </button>
            {/if}
        </div>
    </div>
    {#if showResources}
        <div class="resources-graphs">
            {#if hasCpuLimit}
                {@const cpuLimit = parseCPU(
                    container.DeployResources.CPULimit || "0",
                )}
                {@const cpuReservation = parseCPU(
                    container.DeployResources.CPUReservation || "0",
                )}
                {@const cpuMax = Math.max(
                    cpuLimit || cpuReservation,
                    cpuReservation || cpuLimit,
                    1,
                )}
                {@const reservationPercent =
                    cpuMax > 0 && cpuReservation > 0
                        ? (cpuReservation / cpuMax) * 100
                        : 0}
                {@const limitPercent =
                    cpuMax > 0 && cpuLimit > 0 ? (cpuLimit / cpuMax) * 100 : 0}
                {@const cpuUsage = stats ? stats.cpu : 0}
                {@const usagePercent =
                    cpuMax > 0 && cpuUsage > 0 ? (cpuUsage / cpuMax) * 100 : 0}
                {@const usagePercentCapped = Math.min(usagePercent, 100)}
                <div class="resource-item">
                    <div class="resource-label">
                        <span class="resource-title">CPU</span>
                        <span class="resource-value">
                            {cpuUsage > 0 && cpuLimit > 0
                                ? `${cpuUsage.toFixed(1)} / ${cpuLimit.toFixed(1)} cores (${usagePercent.toFixed(1)}%)`
                                : cpuReservation > 0 && cpuLimit > 0
                                  ? `${cpuReservation.toFixed(1)} / ${cpuLimit.toFixed(1)} cores`
                                  : cpuReservation > 0
                                    ? `${cpuReservation.toFixed(1)} cores`
                                    : cpuLimit.toFixed(1) + " cores"}
                        </span>
                    </div>
                    <div class="progress-bar-container">
                        <div class="progress-bar">
                            {#if cpuReservation > 0}
                                <div
                                    class="progress-fill reservation"
                                    style="width: {reservationPercent}%"
                                ></div>
                            {/if}
                            {#if cpuLimit > 0}
                                <div
                                    class="progress-fill limit"
                                    style="width: {limitPercent}%"
                                ></div>
                            {/if}
                            {#if cpuUsage > 0 && cpuMax > 0}
                                <div
                                    class="progress-fill usage"
                                    style="width: {usagePercentCapped}%"
                                ></div>
                            {/if}
                        </div>
                    </div>
                </div>
            {:else if hasCpuStats}
                {@const cpuUsage = stats ? stats.cpu : 0}
                <div class="resource-item">
                    <div class="resource-label">
                        <span class="resource-title">CPU</span>
                        <span class="resource-value">
                            {cpuUsage.toFixed(1)} cores
                        </span>
                    </div>
                    <div class="progress-bar-container">
                        <div class="progress-bar no-limit">
                            <div
                                class="progress-fill no-limit-fill"
                                style="width: 100%"
                            ></div>
                        </div>
                    </div>
                </div>
            {/if}
            {#if hasMemLimit}
                {@const memLimitMB = parseMemory(
                    container.DeployResources.MemoryLimit || "0",
                )}
                {@const memReservationMB = parseMemory(
                    container.DeployResources.MemoryReservation || "0",
                )}
                {@const memMax = Math.max(
                    memLimitMB || memReservationMB,
                    memReservationMB || memLimitMB,
                    1,
                )}
                {@const reservationPercent =
                    memMax > 0 && memReservationMB > 0
                        ? (memReservationMB / memMax) * 100
                        : 0}
                {@const limitPercent =
                    memMax > 0 && memLimitMB > 0
                        ? (memLimitMB / memMax) * 100
                        : 0}
                {@const memUsageBytes = stats ? stats.memory : 0}
                {@const memUsageMB = memUsageBytes / (1024 * 1024)}
                {@const usagePercent =
                    memMax > 0 && memUsageMB > 0
                        ? (memUsageMB / memMax) * 100
                        : 0}
                {@const usagePercentCapped = Math.min(usagePercent, 100)}
                <div class="resource-item">
                    <div class="resource-label">
                        <span class="resource-title">Memory</span>
                        <span class="resource-value">
                            {memUsageMB > 0 && memLimitMB > 0
                                ? `${formatMemory(memUsageMB)} / ${formatMemory(memLimitMB)} (${usagePercent.toFixed(1)}%)`
                                : memReservationMB > 0 && memLimitMB > 0
                                  ? `${formatMemory(memReservationMB)} / ${formatMemory(memLimitMB)}`
                                  : memReservationMB > 0
                                    ? formatMemory(memReservationMB)
                                    : formatMemory(memLimitMB)}
                        </span>
                    </div>
                    <div class="progress-bar-container">
                        <div class="progress-bar">
                            {#if memReservationMB > 0}
                                <div
                                    class="progress-fill reservation"
                                    style="width: {reservationPercent}%"
                                ></div>
                            {/if}
                            {#if memLimitMB > 0}
                                <div
                                    class="progress-fill limit"
                                    style="width: {limitPercent}%"
                                ></div>
                            {/if}
                            {#if memUsageMB > 0 && memMax > 0}
                                <div
                                    class="progress-fill usage"
                                    style="width: {usagePercentCapped}%"
                                ></div>
                            {/if}
                        </div>
                    </div>
                </div>
            {:else if hasMemStats}
                {@const memUsageBytes = stats ? stats.memory : 0}
                {@const memUsageMB = memUsageBytes / (1024 * 1024)}
                <div class="resource-item">
                    <div class="resource-label">
                        <span class="resource-title">Memory</span>
                        <span class="resource-value">
                            {formatMemory(memUsageMB)}
                        </span>
                    </div>
                    <div class="progress-bar-container">
                        <div class="progress-bar no-limit">
                            <div
                                class="progress-fill no-limit-fill"
                                style="width: 100%"
                            ></div>
                        </div>
                    </div>
                </div>
            {/if}
        </div>
    {/if}
    <p><strong>ID:</strong> {container.ID}</p>
    <p><strong>Image:</strong> {container.Image}</p>
    <p><strong>tag|commit:</strong> {container.TagCommit}</p>
    <p>
        <strong>create image:</strong>
        {container.ImageCreatedAt}
    </p>
    <p>
        <strong>create container:</strong>
        {container.CreatedAt}
    </p>
    <p><strong>uptime container:</strong> {container.Uptime}</p>
    <p><strong>status:</strong> {container.State}</p>
    {#if container.Health}
        <p><strong>health:</strong> {container.Health}</p>
    {/if}
    <p><strong>running:</strong> {container.Run}</p>
    <p><strong>restart:</strong> {container.Restart}</p>
    {#if container.Labels && Object.keys(container.Labels).length > 0}
        <div class="labels-container">
            <strong>Labels:</strong>
            <ul class="labels-list">
                {#each Object.entries(container.Labels) as [key, value]}
                    <li class="label-item">
                        <strong>{key}:</strong>
                        <span class="label-value">{value}</span>
                    </li>
                {/each}
            </ul>
        </div>
    {/if}
</div>

<style>
    .card {
        background-color: white;
        border-radius: 8px;
        box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
        margin: 1rem;
        padding: 1rem;
        width: 300px;
        border: 5px solid transparent;
        position: relative;
    }

    .card.unhealthy {
        border-color: #f44336;
        box-shadow: 0 2px 8px rgba(244, 67, 54, 0.3);
    }

    .card.stopped {
        background-color: #f5f5f5;
        opacity: 0.7;
        border-color: #9e9e9e;
        box-shadow: 0 2px 5px rgba(158, 158, 158, 0.2);
    }

    .card.stopped .card-header h2 {
        color: #757575;
    }

    .card.stopped p {
        color: #9e9e9e;
    }

    .card.stopped .labels-container {
        color: #9e9e9e;
    }

    .card.stopped .labels-container strong {
        color: #757575;
    }

    .card.stopped .label-item {
        color: #9e9e9e;
    }

    .card.stopped .label-item strong {
        color: #757575;
    }

    .card.stopped .label-value {
        color: #9e9e9e;
    }

    .resources-graphs {
        margin: 0.5rem 0;
        padding: 0.5rem;
        background-color: #f8f9fa;
        border-radius: 4px;
        border: 1px solid #e0e0e0;
    }

    .resource-item {
        margin-bottom: 0.5rem;
    }

    .resource-item:last-child {
        margin-bottom: 0;
    }

    .resource-label {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.25rem;
        font-size: 0.85rem;
    }

    .resource-title {
        font-weight: 600;
        color: #333;
    }

    .progress-bar-container {
        margin-top: 0.15rem;
    }

    .progress-bar {
        position: relative;
        width: 100%;
        height: 10px;
        background-color: #e0e0e0;
        border-radius: 5px;
        overflow: hidden;
    }

    .progress-bar.no-limit {
        background-color: transparent;
        border: 1px solid #9e9e9e;
        border-radius: 5px;
    }

    .progress-fill {
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        border-radius: 5px;
        transition: width 0.3s ease;
    }

    .progress-fill.reservation {
        background-color: #4caf50;
        z-index: 1;
    }

    .progress-fill.limit {
        background-color: #2196f3;
        opacity: 0.5;
        z-index: 2;
        border-right: 2px solid #1976d2;
    }

    .progress-fill.usage {
        background-color: #4caf50;
        z-index: 3;
        opacity: 0.9;
    }

    .progress-fill.no-limit-fill {
        background-color: #ffffff;
        z-index: 1;
    }

    .resource-value {
        font-size: 0.8rem;
        color: #555;
        font-weight: normal;
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    .card-header h2 {
        margin: 0;
        flex: 1;
        font-size: 1.5rem;
    }

    .card-header-buttons {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }

    .logs-button {
        background-color: #4caf50;
        color: white;
        border: none;
        padding: 0.4rem 0.8rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.85rem;
        font-weight: 600;
        transition: background-color 0.2s;
    }

    .logs-button:hover {
        background-color: #45a049;
    }

    .restart-button {
        background-color: #2196f3;
        color: white;
        border: none;
        padding: 0.4rem 0.8rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.85rem;
        font-weight: 600;
        transition: background-color 0.2s;
    }

    .restart-button:hover {
        background-color: #1976d2;
    }

    .card p {
        margin: 0.5rem 0;
    }

    .labels-container {
        margin-top: 0.5rem;
        word-wrap: break-word;
        overflow-wrap: break-word;
    }

    .labels-list {
        margin: 0.5rem 0 0 0;
        padding-left: 1.5rem;
        list-style-type: disc;
    }

    .label-item {
        margin: 0.25rem 0;
        word-wrap: break-word;
        overflow-wrap: break-word;
        word-break: break-all;
    }

    .label-value {
        word-wrap: break-word;
        overflow-wrap: break-word;
        word-break: break-all;
        font-family: monospace;
        font-size: 0.9em;
    }
</style>
