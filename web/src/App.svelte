<script>
  import { onMount, onDestroy } from "svelte";

  let containers = [];
  let containersData = null;
  let loading = true;
  let hostinfo = null;
  let ws = null;
  let wsHostInfo = null;
  let wsStats = null;
  let containerStats = new Map(); // Map<containerID, {cpu: number, memory: number}>
  let containerGroups = [];
  let filterText = "";
  let selectedGroup = null; // null означает "все группы"
  let logsShow = false;
  let containerRestart = false;

  // Модальное окно для логов
  let logsModalOpen = false;
  let logsContainerId = "";
  let logsContainerName = "";
  let logs = [];
  let wsLogs = null;
  let logsContainerRef = null;
  let autoScrollEnabled = true;

  // Версия приложения (build-time переменная)
  const appVersion = __APP_VERSION__;

  // Высота статичных элементов
  let headerHeight = 0;
  let metricsBarHeight = 0;
  let filterBarHeight = 0;
  let groupTagsBarHeight = 0;
  let totalFixedHeight = 0;

  // Состояния для сворачивания блоков на мобильных
  let isMobile = false;
  let metricsRowExpanded = false;
  let groupTagsExpanded = false;

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

  function formatTime(dateString) {
    if (!dateString) return "";
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) {
        console.warn("Invalid date:", dateString);
        return "";
      }
      return date.toLocaleTimeString("ru-RU", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
    } catch (error) {
      console.error("Error formatting time:", error, dateString);
      return "";
    }
  }

  // Парсинг CPU из строки типа "2.0"
  function parseCPU(cpuString) {
    if (!cpuString) return 0;
    const match = cpuString.match(/^([\d.]+)/);
    return match ? parseFloat(match[1]) : 0;
  }

  // Парсинг Memory из строки типа "4G", "2.0G", "512M"
  function parseMemory(memoryString) {
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
  }

  // Форматирование Memory обратно в строку
  function formatMemory(mb) {
    if (mb >= 1024) {
      return (mb / 1024).toFixed(1) + " GB";
    }
    return mb.toFixed(1) + " MB";
  }

  function connectWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log("WebSocket connected");
      loading = false;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        containersData = data;
        containers = data.containers || [];
        // Используем группы из API, если они есть, иначе создаем одну группу из всех контейнеров
        if (data.groups && data.groups.length > 0) {
          containerGroups = data.groups;
        } else {
          // Fallback для обратной совместимости
          containerGroups = [
            {
              project_name: "",
              containers: containers,
            },
          ];
        }
        // Обновляем значение logs_show из API
        if (typeof data.logs_show !== "undefined") {
          logsShow = data.logs_show;
        }
        // Обновляем значение container_restart из API
        if (typeof data.container_restart !== "undefined") {
          containerRestart = data.container_restart;
        }
        loading = false;
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      loading = false;
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected, reconnecting...");
      loading = true;
      // Переподключение через 2 секунды
      setTimeout(connectWebSocket, 2000);
    };
  }

  function connectHostInfoWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/hostinfo`;

    wsHostInfo = new WebSocket(wsUrl);

    wsHostInfo.onopen = () => {
      console.log("HostInfo WebSocket connected");
    };

    wsHostInfo.onmessage = (event) => {
      try {
        hostinfo = JSON.parse(event.data);
      } catch (error) {
        console.error("Error parsing hostinfo WebSocket message:", error);
      }
    };

    wsHostInfo.onerror = (error) => {
      console.error("HostInfo WebSocket error:", error);
    };

    wsHostInfo.onclose = () => {
      console.log("HostInfo WebSocket disconnected, reconnecting...");
      // Переподключение через 2 секунды
      setTimeout(connectHostInfoWebSocket, 2000);
    };
  }

  function connectStatsWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers/stats`;

    wsStats = new WebSocket(wsUrl);

    wsStats.onopen = () => {
      console.log("Stats WebSocket connected");
    };

    wsStats.onmessage = (event) => {
      try {
        const stats = JSON.parse(event.data);
        // Обновляем Map с метриками
        const newStats = new Map();
        if (Array.isArray(stats)) {
          stats.forEach((stat) => {
            newStats.set(stat.ID, {
              cpu: stat.CPUUsage || 0,
              memory: stat.MemoryUsage || 0,
            });
          });
        }
        containerStats = newStats;
      } catch (error) {
        console.error("Error parsing stats WebSocket message:", error);
      }
    };

    wsStats.onerror = (error) => {
      console.error("Stats WebSocket error:", error);
    };

    wsStats.onclose = () => {
      console.log("Stats WebSocket disconnected, reconnecting...");
      // Переподключение через 2 секунды
      setTimeout(connectStatsWebSocket, 2000);
    };
  }

  function openLogsModal(containerId, containerName) {
    logsContainerId = containerId;
    logsContainerName = containerName;
    logs = [];
    autoScrollEnabled = true; // Сбрасываем в true при открытии
    logsModalOpen = true;
    connectLogsWebSocket();
  }

  function isScrolledToBottom(element) {
    if (!element) return true;
    const threshold = 10; // Небольшой порог для учета погрешности
    return (
      element.scrollHeight - element.scrollTop - element.clientHeight <
      threshold
    );
  }

  function scrollToBottom() {
    if (logsContainerRef) {
      logsContainerRef.scrollTop = logsContainerRef.scrollHeight;
    }
  }

  function toggleAutoScroll() {
    autoScrollEnabled = !autoScrollEnabled;
    if (autoScrollEnabled) {
      scrollToBottom();
    }
  }

  function handleLogsScroll() {
    // Просто отслеживаем прокрутку, логика автоскролла в onmessage
  }

  function handleEscapeKey(event) {
    if (event.key === "Escape" && logsModalOpen) {
      closeLogsModal();
    }
  }

  function handleModalContentClick(event) {
    event.stopPropagation();
  }

  function closeLogsModal() {
    logsModalOpen = false;
    if (wsLogs) {
      wsLogs.close();
      wsLogs = null;
    }
    logs = [];
    logsContainerId = "";
    logsContainerName = "";
  }

  function connectLogsWebSocket() {
    if (!logsContainerId) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers/${logsContainerId}/logs`;

    wsLogs = new WebSocket(wsUrl);

    wsLogs.onopen = () => {
      console.log("Logs WebSocket connected");
    };

    wsLogs.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.log) {
          const wasAtBottom = isScrolledToBottom(logsContainerRef);
          logs = [...logs, data.log];
          // Автоскролл только если включен и пользователь был внизу
          if (autoScrollEnabled && wasAtBottom) {
            setTimeout(() => {
              scrollToBottom();
            }, 10);
          }
        } else if (data.error) {
          const wasAtBottom = isScrolledToBottom(logsContainerRef);
          logs = [...logs, `ERROR: ${data.error}`];
          if (autoScrollEnabled && wasAtBottom) {
            setTimeout(() => {
              scrollToBottom();
            }, 10);
          }
        }
      } catch (error) {
        console.error("Error parsing logs WebSocket message:", error);
      }
    };

    wsLogs.onerror = (error) => {
      console.error("Logs WebSocket error:", error);
      logs = [...logs, "ERROR: Failed to connect to logs stream"];
    };

    wsLogs.onclose = () => {
      console.log("Logs WebSocket disconnected");
    };
  }

  // Функция для определения мобильного устройства
  const checkMobile = () => {
    isMobile = window.matchMedia("(max-width: 768px)").matches;
    // На мобильных по умолчанию блоки свернуты
    if (isMobile) {
      metricsRowExpanded = false;
      groupTagsExpanded = false;
    } else {
      metricsRowExpanded = true;
      groupTagsExpanded = true;
    }
  };

  // Обработчики для переключения состояния блоков
  const handleToggleMetricsRow = () => {
    metricsRowExpanded = !metricsRowExpanded;
    // Обновляем высоту после изменения состояния
    setTimeout(() => {
      const metricsBarEl = document.querySelector(".metrics-bar");
      if (metricsBarEl) metricsBarHeight = metricsBarEl.offsetHeight;
      totalFixedHeight =
        headerHeight + metricsBarHeight + filterBarHeight + groupTagsBarHeight;
    }, 350); // Ждем завершения анимации
  };

  const handleToggleGroupTags = () => {
    groupTagsExpanded = !groupTagsExpanded;
    // Обновляем высоту после изменения состояния
    setTimeout(() => {
      const groupTagsBarEl = document.querySelector(".group-tags-bar");
      if (groupTagsBarEl) groupTagsBarHeight = groupTagsBarEl.offsetHeight;
      totalFixedHeight =
        headerHeight + metricsBarHeight + filterBarHeight + groupTagsBarHeight;
    }, 350); // Ждем завершения анимации
  };

  onMount(() => {
    connectWebSocket();
    connectHostInfoWebSocket();
    connectStatsWebSocket();
    window.addEventListener("keydown", handleEscapeKey);
    checkMobile();

    // Вычисляем высоту статичных элементов
    const updateFixedHeights = () => {
      const headerEl = document.querySelector("header");
      const metricsBarEl = document.querySelector(".metrics-bar");
      const filterBarEl = document.querySelector(".filter-bar");
      const groupTagsBarEl = document.querySelector(".group-tags-bar");

      headerHeight = headerEl ? headerEl.offsetHeight : 0;
      metricsBarHeight = metricsBarEl ? metricsBarEl.offsetHeight : 0;
      filterBarHeight = filterBarEl ? filterBarEl.offsetHeight : 0;
      groupTagsBarHeight = groupTagsBarEl ? groupTagsBarEl.offsetHeight : 0;
      totalFixedHeight =
        headerHeight + metricsBarHeight + filterBarHeight + groupTagsBarHeight;
    };

    updateFixedHeights();
    // Обновляем при изменении размеров окна
    const handleResize = () => {
      checkMobile();
      updateFixedHeights();
    };
    window.addEventListener("resize", handleResize);

    // Следим за изменением медиа-запроса
    const mediaQuery = window.matchMedia("(max-width: 768px)");
    const handleMediaChange = (e) => {
      checkMobile();
      updateFixedHeights();
    };
    mediaQuery.addEventListener("change", handleMediaChange);
  });

  // Обновляем высоту при изменении hostinfo, групп или состояния свернутости
  $: if (
    hostinfo ||
    allGroups.length > 0 ||
    metricsRowExpanded !== undefined ||
    groupTagsExpanded !== undefined
  ) {
    setTimeout(() => {
      const headerEl = document.querySelector("header");
      const metricsBarEl = document.querySelector(".metrics-bar");
      const filterBarEl = document.querySelector(".filter-bar");
      const groupTagsBarEl = document.querySelector(".group-tags-bar");

      if (headerEl) headerHeight = headerEl.offsetHeight;
      if (metricsBarEl) metricsBarHeight = metricsBarEl.offsetHeight;
      if (filterBarEl) filterBarHeight = filterBarEl.offsetHeight;
      if (groupTagsBarEl) groupTagsBarHeight = groupTagsBarEl.offsetHeight;
      totalFixedHeight =
        headerHeight + metricsBarHeight + filterBarHeight + groupTagsBarHeight;
    }, 100);
  }

  onDestroy(() => {
    if (ws) {
      ws.close();
    }
    if (wsHostInfo) {
      wsHostInfo.close();
    }
    if (wsStats) {
      wsStats.close();
    }
    if (wsLogs) {
      wsLogs.close();
    }
    window.removeEventListener("keydown", handleEscapeKey);
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

  // Получение списка всех уникальных групп
  $: allGroups = (() => {
    const groups = new Set();
    containerGroups.forEach((group) => {
      if (group.project_name !== undefined) {
        groups.add(group.project_name || ""); // Пустая строка для группы без проекта
      }
    });
    return Array.from(groups).sort((a, b) => {
      if (a === "") return 1; // Группа без проекта в конец
      if (b === "") return -1;
      return a.localeCompare(b);
    });
  })();

  // Фильтрация групп контейнеров по имени и выбранной группе
  $: filteredGroups = (() => {
    // Сначала фильтруем по выбранной группе
    let groupsToFilter = containerGroups;
    if (selectedGroup !== null) {
      groupsToFilter = containerGroups.filter((group) => {
        const groupName = group.project_name || "";
        return groupName === selectedGroup;
      });
    }

    // Затем применяем фильтр по имени контейнера
    if (!filterText.trim()) {
      return groupsToFilter;
    }

    const filterLower = filterText.toLowerCase().trim();
    return groupsToFilter
      .map((group) => {
        const filteredContainers = group.containers.filter((container) =>
          container.Name.toLowerCase().includes(filterLower),
        );

        if (filteredContainers.length === 0) {
          return null;
        }

        return {
          ...group,
          containers: filteredContainers,
        };
      })
      .filter((group) => group !== null);
  })();

  // Обработчик клика по тегу группы
  const handleGroupClick = (groupName) => {
    if (selectedGroup === groupName) {
      selectedGroup = null; // Снимаем фильтр при повторном клике
    } else {
      selectedGroup = groupName;
    }
  };

  // Обработчик перезагрузки контейнера через WebSocket
  const handleRestartContainer = (containerId, containerName) => {
    if (!containerRestart) {
      console.warn("Container restart is disabled");
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers/${containerId}/restart`;

    const wsRestart = new WebSocket(wsUrl);

    wsRestart.onopen = () => {
      console.log("Restart WebSocket connected");
    };

    wsRestart.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.status === "success") {
          console.log(
            `Container ${containerName} restarted successfully:`,
            data.message,
          );
        } else {
          console.error(
            `Failed to restart container ${containerName}:`,
            data.message,
          );
          alert(
            `Failed to restart container ${containerName}: ${data.message}`,
          );
        }
      } catch (error) {
        console.error("Error parsing restart response:", error);
        alert(`Failed to restart container ${containerName}: Invalid response`);
      }
      wsRestart.close();
    };

    wsRestart.onerror = (error) => {
      console.error(`WebSocket error for restart ${containerName}:`, error);
      alert(`Failed to restart container ${containerName}: Connection error`);
      wsRestart.close();
    };

    wsRestart.onclose = () => {
      console.log("Restart WebSocket disconnected");
    };
  };
</script>

<div class="scrollable-content">
  <header>
    <div class="header-left">
      <svg
        class="logo"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" />
        <path
          d="M2 17L12 22L22 17"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          fill="none"
        />
        <path
          d="M2 12L12 17L22 12"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          fill="none"
        />
      </svg>
      <h1>CoreOps Dashboard</h1>
    </div>
    <div class="header-right">
      <span class="version-text">{appVersion}</span>
      <a
        href="https://github.com/SSomov/docker-dashboard"
        target="_blank"
        rel="noopener noreferrer"
        class="header-link"
        aria-label="GitHub Repository"
      >
        <svg
          viewBox="0 0 24 24"
          fill="currentColor"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
          />
        </svg>
      </a>
      <a
        href="https://hub.docker.com/r/avt0x/docker-dashboard"
        target="_blank"
        rel="noopener noreferrer"
        class="header-link"
        aria-label="Docker Hub"
      >
        <svg
          class="docker-hub-logo"
          viewBox="0 0 24 24"
          fill="white"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M23.6096 7.78281C23.0211 7.38668 21.4751 7.21752 20.3513 7.52051C20.291 6.40098 19.7132 5.45739 18.6572 4.63407L18.2664 4.37177L18.0059 4.7654C17.4937 5.54269 17.2778 6.5787 17.3542 7.52051C17.4145 8.10079 17.6162 8.75281 18.0059 9.22603C16.543 10.0747 15.1944 9.88197 9.22236 9.88197H0.00204151C-0.0247244 11.2306 0.191901 13.8244 1.84139 15.9361C2.02376 16.1695 2.22326 16.395 2.44024 16.6124C3.78139 17.9553 5.80775 18.9403 8.83765 18.9428C13.4603 18.9471 17.4206 16.4482 19.8299 10.4069C20.6225 10.4198 22.7152 10.549 23.7395 8.57009C23.7645 8.5369 24 8.04548 24 8.04548L23.6096 7.78317V7.78281ZM6.01938 6.5498H3.42665V9.14252H6.01938V6.5498ZM9.36868 6.5498H6.77596V9.14252H9.36868V6.5498ZM12.7183 6.5498H10.1256V9.14252H12.7183V6.5498ZM16.068 6.5498H13.4753V9.14252H16.068V6.5498ZM2.66971 6.5498H0.0769861V9.14252H2.66971V6.5498ZM6.01938 3.27508H3.42665V5.8678H6.01938V3.27508ZM9.36868 3.27508H6.77596V5.8678H9.36868V3.27508ZM12.7183 3.27508H10.1256V5.8678H12.7183V3.27508ZM12.7183 0H10.1256V2.59272H12.7183V0Z"
            fill="white"
          />
        </svg>
      </a>
    </div>
  </header>

  {#if hostinfo}
    <div class="metrics-bar" style="top: {headerHeight}px">
      <div class="metrics-header">
        <div class="metrics-header-left">
          <span><b>Host:</b> {hostinfo.host.hostname}</span>
          <span><b>Uptime:</b> {(hostinfo.host.uptime / 3600).toFixed(1)}h</span
          >
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
              on:click={handleToggleMetricsRow}
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

  <div class="filter-bar" style="top: {headerHeight + metricsBarHeight}px">
    <label for="container-filter">Filter:</label>
    <input
      id="container-filter"
      type="text"
      placeholder="Filter containers by name..."
      bind:value={filterText}
    />
  </div>

  {#if allGroups.length > 0}
    <div
      class="group-tags-bar"
      style="top: {headerHeight + metricsBarHeight + filterBarHeight}px"
    >
      {#if isMobile}
        <div class="group-tags-header">
          <span class="group-tags-title">Groups</span>
          <button
            class="toggle-button"
            on:click={handleToggleGroupTags}
            aria-label={groupTagsExpanded
              ? "Свернуть группы"
              : "Развернуть группы"}
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class:rotated={groupTagsExpanded}
            >
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
          </button>
        </div>
      {/if}
      <div
        class="group-tags-container"
        class:hidden-mobile={isMobile && !groupTagsExpanded}
      >
        <button
          class="group-tag"
          class:active={selectedGroup === null}
          on:click={() => handleGroupClick(null)}
        >
          All
        </button>
        {#each allGroups as groupName (groupName)}
          <button
            class="group-tag"
            class:active={selectedGroup === groupName}
            on:click={() => handleGroupClick(groupName)}
          >
            {groupName || "No Project"}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <div class="content-wrapper" style="padding-top: {totalFixedHeight}px">
    {#if loading}
      <div class="container">
        <div class="loader-container">
          <div class="loader"></div>
        </div>
      </div>
    {:else}
      <div class="container">
        {#each filteredGroups as group (group.project_name || "ungrouped")}
          <div class="container-group">
            {#if group.project_name}
              <div class="group-header">
                Project: {group.project_name}
              </div>
            {/if}
            <div class="group-containers">
              {#each group.containers as container (container.ID)}
                {@const stats = containerStats.get(container.ID)}
                {@const hasCpuLimit = container.DeployResources && (container.DeployResources.CPULimit || container.DeployResources.CPUReservation)}
                {@const hasMemLimit = container.DeployResources && (container.DeployResources.MemoryLimit || container.DeployResources.MemoryReservation)}
                {@const hasCpuStats = stats && stats.cpu > 0}
                {@const hasMemStats = stats && stats.memory > 0}
                {@const showResources = hasCpuLimit || hasMemLimit || hasCpuStats || hasMemStats}
                <div
                  class="card"
                  class:unhealthy={container.Health === "unhealthy"}
                  class:stopped={!container.Run ||
                    container.State !== "running"}
                >
                  <div class="card-header">
                    <h2>{container.Name}</h2>
                    <div class="card-header-buttons">
                      {#if logsShow}
                        <button
                          class="logs-button"
                          on:click={() =>
                            openLogsModal(container.ID, container.Name)}
                        >
                          logs
                        </button>
                      {/if}
                      {#if containerRestart}
                        <button
                          class="restart-button"
                          on:click={() =>
                            handleRestartContainer(
                              container.ID,
                              container.Name,
                            )}
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
                          cpuMax > 0 && cpuLimit > 0
                            ? (cpuLimit / cpuMax) * 100
                            : 0}
                        {@const cpuUsage = stats ? stats.cpu : 0}
                        {@const usagePercent =
                          cpuMax > 0 && cpuUsage > 0
                            ? (cpuUsage / cpuMax) * 100
                            : 0}
                        {@const usagePercentCapped = Math.min(usagePercent, 100)}
                        <div class="resource-item">
                          <div class="resource-label">
                            <span class="resource-title">CPU</span>
                            <span class="resource-value">
                              {cpuUsage > 0 && cpuLimit > 0
                                ? `${cpuUsage.toFixed(1)} / ${cpuLimit.toFixed(1)} cores (${usagePercent.toFixed(1)}%)`
                                : cpuReservation > 0
                                  ? `${cpuReservation.toFixed(1)} cores`
                                  : cpuLimit.toFixed(1) + " cores"}
                              {cpuLimit > 0 && cpuUsage === 0 ? ` (limit)` : ""}
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
                                : memReservationMB > 0
                                  ? formatMemory(memReservationMB)
                                  : formatMemory(memLimitMB)}
                              {memLimitMB > 0 && memReservationMB > 0 && memUsageMB === 0
                                ? ` / ${formatMemory(memLimitMB)}`
                                : memLimitMB > 0 && memUsageMB === 0
                                  ? ` (limit: ${formatMemory(memLimitMB)})`
                                  : ""}
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
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Модальное окно для логов -->
{#if logsModalOpen}
  <div
    class="modal-overlay"
    role="button"
    tabindex="0"
    on:click={closeLogsModal}
    on:keydown={(e) => e.key === "Enter" && closeLogsModal()}
  >
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal-content"
      role="dialog"
      aria-labelledby="modal-title"
      on:click={handleModalContentClick}
    >
      <div class="modal-header">
        <h2 id="modal-title">Logs: {logsContainerName}</h2>
        <div class="modal-header-controls">
          <label class="auto-scroll-toggle">
            <input
              type="checkbox"
              bind:checked={autoScrollEnabled}
              on:change={toggleAutoScroll}
            />
            <span>Auto-scroll</span>
          </label>
          <button
            class="modal-close"
            on:click={closeLogsModal}
            aria-label="Close logs modal">×</button
          >
        </div>
      </div>
      <div
        class="modal-body"
        bind:this={logsContainerRef}
        on:scroll={handleLogsScroll}
      >
        {#each logs as log, index (index)}
          <div class="log-line">{log}</div>
        {/each}
        {#if logs.length === 0}
          <div class="log-line log-empty">Waiting for logs...</div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    font-family: Arial, sans-serif;
    background-color: #234255;
    color: #333;
    margin: 0;
    padding: 0;
  }

  header {
    background-color: #4caf50;
    color: white;
    padding: 0.4rem 1rem;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  header h1 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
  }

  .logo {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .header-link {
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    text-decoration: none;
    transition: opacity 0.2s;
    flex-shrink: 0;
  }

  .header-link:hover {
    opacity: 0.7;
  }

  .header-link svg {
    width: 24px;
    height: 24px;
  }

  .docker-hub-logo {
    width: 120px;
    height: 37px;
    display: block;
    margin: 0;
    padding: 0;
  }

  .header-link:has(.docker-hub-logo) {
    padding: 0;
    margin: 0;
    width: auto;
    height: auto;
  }

  .version-text {
    color: white;
    font-size: 0.9rem;
    font-weight: 500;
    white-space: nowrap;
    opacity: 0.9;
  }

  .scrollable-content {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    overflow-y: auto;
    overflow-x: hidden;
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

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    pointer-events: auto;
  }

  .modal-content {
    background-color: white;
    border-radius: 8px;
    width: 90%;
    max-width: 900px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    border-bottom: 2px solid #e0e0e0;
    background-color: #f5f5f5;
    border-radius: 8px 8px 0 0;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.2rem;
    color: #333;
    flex: 1;
  }

  .modal-header-controls {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .auto-scroll-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    user-select: none;
    font-size: 0.9rem;
    color: #333;
  }

  .auto-scroll-toggle input[type="checkbox"] {
    cursor: pointer;
    width: 18px;
    height: 18px;
  }

  .auto-scroll-toggle span {
    white-space: nowrap;
  }

  .modal-close {
    background: none;
    border: none;
    font-size: 2rem;
    cursor: pointer;
    color: #666;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: background-color 0.2s;
  }

  .modal-close:hover {
    background-color: #e0e0e0;
  }

  .modal-body {
    padding: 1rem;
    overflow-y: auto;
    flex: 1;
    background-color: #1e1e1e;
    color: #d4d4d4;
    font-family: "Courier New", monospace;
    font-size: 0.9rem;
    line-height: 1.5;
  }

  .log-line {
    margin: 0;
    padding: 0.2rem 0;
    word-wrap: break-word;
    white-space: pre-wrap;
  }

  .log-empty {
    color: #888;
    font-style: italic;
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

  .filter-bar {
    background: #2c3e50;
    padding: 0.75rem 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: fixed;
    left: 0;
    right: 0;
    z-index: 99;
  }

  .filter-bar label {
    color: #fff;
    font-weight: 600;
    white-space: nowrap;
  }

  .filter-bar input {
    flex: 1;
    padding: 0.5rem 1rem;
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    font-size: 0.95rem;
  }

  .filter-bar input::placeholder {
    color: rgba(255, 255, 255, 0.5);
  }

  .filter-bar input:focus {
    outline: none;
    border-color: #4caf50;
    background: rgba(255, 255, 255, 0.15);
  }

  .group-tags-bar {
    background: #34495e;
    padding: 0.75rem 1.5rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: fixed;
    left: 0;
    right: 0;
    z-index: 99;
  }

  .group-tags-header {
    display: none;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  }

  @media (max-width: 768px) {
    .group-tags-header {
      display: flex;
    }
  }

  .group-tags-title {
    color: #fff;
    font-weight: 600;
    font-size: 0.95rem;
  }

  .group-tags-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
    transition:
      max-height 0.3s ease,
      opacity 0.3s ease;
    overflow: hidden;
  }

  .group-tags-container.hidden-mobile {
    max-height: 0;
    opacity: 0;
  }

  .group-tag {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 4px;
    padding: 0.4rem 0.8rem;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    white-space: nowrap;
  }

  .group-tag:hover {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.5);
  }

  .group-tag.active {
    background: #4caf50;
    border-color: #4caf50;
    color: white;
    font-weight: 600;
    box-shadow: 0 2px 4px rgba(76, 175, 80, 0.3);
  }

  .group-tag.active:hover {
    background: #45a049;
    border-color: #45a049;
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

  .group-containers .card {
    margin: 1rem;
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
    .metrics-row.hidden-mobile,
    .group-tags-container.hidden-mobile {
      max-height: none !important;
      opacity: 1 !important;
      margin-top: 0.5rem !important;
    }
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
