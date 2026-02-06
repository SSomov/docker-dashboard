<script>
import { createEventDispatcher } from "svelte";
import {
	formatTime,
	getCpuPercent,
	getRamPercent,
	getSwapPercent,
} from "../utils/formatters.js";

export let hostinfo = null;
export let containersData = null;
export let isMobile = false;
export let metricsRowExpanded = false;
export let headerHeight = 0;

const dispatch = createEventDispatcher();

$: totalDisk = (() => {
	if (!hostinfo || !hostinfo.disk_usage) return { used: 0, total: 0 };
	let used = 0,
		total = 0;
	for (const mount in hostinfo.disk_usage) {
		const usage = hostinfo.disk_usage[mount];
		used += usage.used;
		total += usage.total;
	}
	return { used, total };
})();

function handleToggle() {
	dispatch("toggle");
}
</script>

{#if hostinfo}
  <div class="metrics-bar" style="top: {headerHeight}px">
    <div class="metrics-header">
      <div class="metrics-header-left">
        <span><b>Host:</b> {hostinfo.host.hostname}</span>
        <span><b>Uptime:</b> {(hostinfo.host.uptime / 3600).toFixed(1)}h</span>
      </div>
      <div class="metrics-header-right">
        {#if containersData}
          <span><b>Containers:</b> {containersData.total || 0}</span>
          <span
            ><b>Snapshot:</b>
            {containersData.snapshot_time
              ? formatTime(containersData.snapshot_time)
              : "-"}</span
          >
        {/if}
        {#if isMobile}
          <button
            class="toggle-button"
            on:click={handleToggle}
            aria-label={metricsRowExpanded
              ? "Свернуть метрики"
              : "Развернуть метрики"}
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class:rotated={metricsRowExpanded}
            >
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
          </button>
        {/if}
      </div>
    </div>
    <div
      class="metrics-row"
      class:hidden-mobile={isMobile && !metricsRowExpanded}
    >
      <span>
        <b>CPU:</b>
        {getCpuPercent(hostinfo.cpu).toFixed(1)}%
        <div class="mini-progress">
          <div
            class="mini-progress-bar cpu"
            style="width: {getCpuPercent(hostinfo.cpu)}%"
          ></div>
        </div>
      </span>
      <span>
        <b>Load:</b>
        {hostinfo.load ? hostinfo.load.load1.toFixed(2) : "-"}/
        {hostinfo.load ? hostinfo.load.load5.toFixed(2) : "-"}/
        {hostinfo.load ? hostinfo.load.load15.toFixed(2) : "-"}
      </span>
      <span>
        <b>RAM:</b>
        {(hostinfo.memory.used / 1024 / 1024).toFixed(1)} /
        {(hostinfo.memory.total / 1024 / 1024).toFixed(1)} MiB ({getRamPercent(
          hostinfo.memory.used,
          hostinfo.memory.total,
        ).toFixed(1)}%)
        <div class="mini-progress">
          <div
            class="mini-progress-bar ram"
            style="width: {getRamPercent(
              hostinfo.memory.used,
              hostinfo.memory.total,
            )}%"
          ></div>
        </div>
      </span>
      <span>
        <b>Swap:</b>
        {#if hostinfo.memory && hostinfo.memory.swapTotal > 0}
          {(
            (hostinfo.memory.swapTotal - hostinfo.memory.swapFree) /
            1024 /
            1024
          ).toFixed(2)} /
          {(hostinfo.memory.swapTotal / 1024 / 1024).toFixed(2)} MiB ({getSwapPercent(
            hostinfo.memory.swapTotal,
            hostinfo.memory.swapFree,
          ).toFixed(1)}%)
          <div class="mini-progress">
            <div
              class="mini-progress-bar swap"
              style="width: {getSwapPercent(
                hostinfo.memory.swapTotal,
                hostinfo.memory.swapFree,
              )}%"
            ></div>
          </div>
        {:else}
          no swap
        {/if}
      </span>
      <span>
        <b>Disk:</b>
        {#if totalDisk.total > 0}
          {(totalDisk.used / 1024 / 1024 / 1024).toFixed(2)} /
          {(totalDisk.total / 1024 / 1024 / 1024).toFixed(2)} GiB
          {((totalDisk.used / totalDisk.total) * 100).toFixed(1)}%
          <div class="mini-progress">
            <div
              class="mini-progress-bar disk"
              style="width: {(totalDisk.used / totalDisk.total) * 100}%"
            ></div>
          </div>
        {/if}
      </span>
      <span>
        <b>I/O:</b>
        {#if hostinfo.disk_usage && hostinfo.disk_usage["/"]}
          R {(hostinfo.disk_usage["/"].readBytes
            ? hostinfo.disk_usage["/"].readBytes / 1024 / 1024
            : 0
          ).toFixed(2)} MiB, W {(hostinfo.disk_usage["/"].writeBytes
            ? hostinfo.disk_usage["/"].writeBytes / 1024 / 1024
            : 0
          ).toFixed(2)} MiB
        {/if}
      </span>
      <span>
        <b>Net:</b>
        {#if hostinfo.net && hostinfo.net.length}
          RX {(hostinfo.net[0].bytesRecv
            ? hostinfo.net[0].bytesRecv / 1024 / 1024
            : 0
          ).toFixed(2)} MiB, TX {(hostinfo.net[0].bytesSent
            ? hostinfo.net[0].bytesSent / 1024 / 1024
            : 0
          ).toFixed(2)} MiB
        {/if}
      </span>
    </div>
  </div>
{/if}

<style>
  .metrics-bar {
    background: #2c3e50;
    color: #fff;
    padding: 0.75rem 1.5rem;
    font-size: 0.95rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: fixed;
    left: 0;
    right: 0;
    z-index: 99;
  }

  .metrics-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
    font-weight: 600;
  }

  .metrics-header-left {
    display: flex;
    gap: 1.5rem;
    align-items: center;
  }

  .metrics-header-right {
    display: flex;
    gap: 1.5rem;
    align-items: center;
    font-size: 0.9rem;
    color: rgba(255, 255, 255, 0.9);
  }

  @media (max-width: 768px) {
    .metrics-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
    }

    .metrics-header-left,
    .metrics-header-right {
      width: 100%;
      justify-content: space-between;
      flex-wrap: wrap;
    }
  }

  .metrics-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1.5rem;
    justify-content: flex-start;
    align-items: center;
    margin-top: 0.5rem;
    transition:
      max-height 0.3s ease,
      opacity 0.3s ease;
    overflow: hidden;
  }

  .metrics-row.hidden-mobile {
    max-height: 0;
    margin-top: 0;
    opacity: 0;
  }

  .metrics-bar span {
    margin: 0 0.25rem;
    white-space: nowrap;
    display: flex;
    align-items: center;
    padding: 0.25rem 0.5rem;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }

  .metrics-bar span b {
    margin-right: 0.4rem;
    font-weight: 600;
  }

  .mini-progress {
    width: 70px;
    height: 10px;
    background-color: #ddd;
    border-radius: 4px;
    overflow: hidden;
    margin-left: 0.5em;
    display: inline-block;
  }

  .mini-progress-bar {
    height: 100%;
    transition: width 0.3s ease;
  }

  .mini-progress-bar.cpu {
    background-color: #ff9800;
  }

  .mini-progress-bar.ram {
    background-color: #4caf50;
  }

  .mini-progress-bar.swap {
    background-color: #2196f3;
  }

  .mini-progress-bar.disk {
    background-color: #9c27b0;
  }

  .toggle-button {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 4px;
    color: #fff;
    cursor: pointer;
    padding: 0.4rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
    flex-shrink: 0;
    width: 32px;
    height: 32px;
  }

  .toggle-button:hover {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.5);
  }

  .toggle-button svg {
    width: 20px;
    height: 20px;
    transition: transform 0.3s ease;
  }

  .toggle-button svg.rotated {
    transform: rotate(180deg);
  }

  @media (min-width: 769px) {
    .metrics-row.hidden-mobile {
      max-height: none !important;
      opacity: 1 !important;
      margin-top: 0.5rem !important;
    }
  }
</style>

