<script>
  import { onMount, onDestroy } from 'svelte';
  import axios from 'axios';

  let containers = [];
  let containersData = null;
  let loading = true;
  let hostinfo = null;
  let ws = null;

  $: totalDisk = (() => {
    if (!hostinfo || !hostinfo.disk_usage) return { used: 0, total: 0 };
    let used = 0, total = 0;
    for (const mount in hostinfo.disk_usage) {
      const usage = hostinfo.disk_usage[mount];
      used += usage.used;
      total += usage.total;
    }
    return { used, total };
  })();

  function getBaseUrl() {
    return `${window.location.origin}${window.location.pathname}`;
  }

  function formatTime(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }

  function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers`;
    
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      loading = false;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        containersData = data;
        containers = data.containers || [];
        loading = false;
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      loading = false;
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected, reconnecting...');
      loading = true;
      // Переподключение через 2 секунды
      setTimeout(connectWebSocket, 2000);
    };
  }

  function fetchHostInfo() {
    axios.get(`${getBaseUrl()}/api/hostinfo`)
      .then(response => {
        hostinfo = response.data;
        setTimeout(fetchHostInfo, 2000);
      })
      .catch(error => {
        console.error('Error fetching hostinfo:', error);
        setTimeout(fetchHostInfo, 5000);
      });
  }

  onMount(() => {
    connectWebSocket();
    fetchHostInfo();
  });

  onDestroy(() => {
    if (ws) {
      ws.close();
    }
  });

  function getCpuPercent(cpu) {
    return cpu && cpu.length ? cpu[0] : 0;
  }

  function getRamPercent(used, total) {
    return used && total ? (used / total) * 100 : 0;
  }

  function getSwapPercent(swapTotal, swapFree) {
    return swapTotal > 0 ? ((swapTotal - swapFree) / swapTotal) * 100 : 0;
  }
</script>

<style>
  :global(body) {
    font-family: Arial, sans-serif;
    background-color: #234255;
    color: #333;
    margin: 0;
    padding: 0;
  }

  header {
    background-color: #4CAF50;
    color: white;
    padding: 0.25rem;
    text-align: center;
  }

  .container {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    margin: 2rem;
  }

  .card {
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
    margin: 1rem;
    padding: 1rem;
    width: 300px;
  }

  .card h2 {
    font-size: 1.5rem;
    margin-top: 0;
  }

  .card p {
    margin: 0.5rem 0;
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
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .loader-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
  }

  .metrics-bar {
    background: #2c3e50;
    color: #fff;
    padding: 0.75rem 1.5rem;
    font-size: 0.95rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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

  .metrics-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1.5rem;
    justify-content: flex-start;
    align-items: center;
    margin-top: 0.5rem;
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
</style>

<header>
  <h1>List of running node containers</h1>
</header>

{#if hostinfo}
  <div class="metrics-bar">
    <div class="metrics-header">
      <div class="metrics-header-left">
        <span><b>Host:</b> {hostinfo.host.hostname}</span>
        <span><b>Uptime:</b> {(hostinfo.host.uptime / 3600).toFixed(1)}h</span>
      </div>
      <div class="metrics-header-right">
        {#if containersData}
          <span><b>Containers:</b> {containersData.total}</span>
          <span><b>Snapshot:</b> {formatTime(containersData.snapshot_time)}</span>
        {/if}
      </div>
    </div>
    <div class="metrics-row">
      <span>
        <b>CPU:</b> {getCpuPercent(hostinfo.cpu).toFixed(1)}%
        <div class="mini-progress">
          <div class="mini-progress-bar cpu" style="width: {getCpuPercent(hostinfo.cpu)}%"></div>
        </div>
      </span>
      <span>
        <b>Load:</b> {hostinfo.load ? hostinfo.load.load1.toFixed(2) : '-'}/
        {hostinfo.load ? hostinfo.load.load5.toFixed(2) : '-'}/
        {hostinfo.load ? hostinfo.load.load15.toFixed(2) : '-'}
      </span>
      <span>
        <b>RAM:</b> {(hostinfo.memory.used / 1024 / 1024).toFixed(1)} / 
        {(hostinfo.memory.total / 1024 / 1024).toFixed(1)} MiB 
        ({getRamPercent(hostinfo.memory.used, hostinfo.memory.total).toFixed(1)}%)
        <div class="mini-progress">
          <div class="mini-progress-bar ram" style="width: {getRamPercent(hostinfo.memory.used, hostinfo.memory.total)}%"></div>
        </div>
      </span>
      <span>
        <b>Swap:</b>
        {#if hostinfo.memory && hostinfo.memory.swapTotal > 0}
          {((hostinfo.memory.swapTotal - hostinfo.memory.swapFree) / 1024 / 1024).toFixed(2)} / 
          {(hostinfo.memory.swapTotal / 1024 / 1024).toFixed(2)} MiB 
          ({getSwapPercent(hostinfo.memory.swapTotal, hostinfo.memory.swapFree).toFixed(1)}%)
          <div class="mini-progress">
            <div class="mini-progress-bar swap" style="width: {getSwapPercent(hostinfo.memory.swapTotal, hostinfo.memory.swapFree)}%"></div>
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
            <div class="mini-progress-bar disk" style="width: {(totalDisk.used / totalDisk.total) * 100}%"></div>
          </div>
        {/if}
      </span>
      <span>
        <b>I/O:</b>
        {#if hostinfo.disk_usage && hostinfo.disk_usage['/']}
          R {(hostinfo.disk_usage['/'].readBytes ? hostinfo.disk_usage['/'].readBytes / 1024 / 1024 : 0).toFixed(2)} MiB, 
          W {(hostinfo.disk_usage['/'].writeBytes ? hostinfo.disk_usage['/'].writeBytes / 1024 / 1024 : 0).toFixed(2)} MiB
        {/if}
      </span>
      <span>
        <b>Net:</b>
        {#if hostinfo.net && hostinfo.net.length}
          RX {(hostinfo.net[0].bytesRecv ? hostinfo.net[0].bytesRecv / 1024 / 1024 : 0).toFixed(2)} MiB, 
          TX {(hostinfo.net[0].bytesSent ? hostinfo.net[0].bytesSent / 1024 / 1024 : 0).toFixed(2)} MiB
        {/if}
      </span>
    </div>
  </div>
{/if}

{#if loading}
  <div class="container">
    <div class="loader-container">
      <div class="loader"></div>
    </div>
  </div>
{:else}
  <div class="container">
    {#each containers as container (container.ID)}
      <div class="card">
        <h2>{container.Name}</h2>
        <p><strong>ID:</strong> {container.ID}</p>
        <p><strong>Image:</strong> {container.Image}</p>
        <p><strong>tag|commit:</strong> {container.TagCommit}</p>
        <p><strong>create image:</strong> {container.ImageCreatedAt}</p>
        <p><strong>create container:</strong> {container.CreatedAt}</p>
        <p><strong>uptime container:</strong> {container.Uptime}</p>
        <p><strong>status:</strong> {container.State}</p>
        <p><strong>health:</strong> {container.Health}</p>
        <p><strong>running:</strong> {container.Run}</p>
        <p><strong>restart:</strong> {container.Restart}</p>
        <div>
          <strong>Labels:</strong>
          <ul>
            {#each Object.entries(container.Labels || {}) as [key, value]}
              <li><strong>{key}:</strong> {value}</li>
            {/each}
          </ul>
        </div>
      </div>
    {/each}
  </div>
{/if}

