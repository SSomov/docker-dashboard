<script>
import { onDestroy, onMount } from "svelte";
import ContainerGroups from "./components/ContainerGroups.svelte";
import FilterBar from "./components/FilterBar.svelte";
import GroupTagsBar from "./components/GroupTagsBar.svelte";
import Header from "./components/Header.svelte";
import LogsModal from "./components/LogsModal.svelte";
import MetricsBar from "./components/MetricsBar.svelte";
import { createWebSocketStore } from "./composables/websocket.js";
import { checkMobile, updateFixedHeights } from "./utils/layout.js";

// Версия приложения (build-time переменная)
const appVersion = __APP_VERSION__;

// Состояние приложения
let containers = [];
let containersData = null;
let loading = true;
let hostinfo = null;
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

// WebSocket store
const wsStore = createWebSocketStore();

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

// Функция для определения мобильного устройства
const checkMobileDevice = () => {
	isMobile = checkMobile();
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
		updateHeights();
	}, 350); // Ждем завершения анимации
};

const handleToggleGroupTags = () => {
	groupTagsExpanded = !groupTagsExpanded;
	// Обновляем высоту после изменения состояния
	setTimeout(() => {
		updateHeights();
	}, 350); // Ждем завершения анимации
};

function updateHeights() {
	const heights = updateFixedHeights();
	headerHeight = heights.headerHeight;
	metricsBarHeight = heights.metricsBarHeight;
	filterBarHeight = heights.filterBarHeight;
	groupTagsBarHeight = heights.groupTagsBarHeight;
	totalFixedHeight = heights.totalFixedHeight;
}

function openLogsModal(containerId, containerName) {
	logsContainerId = containerId;
	logsContainerName = containerName;
	logsModalOpen = true;
}

function closeLogsModal() {
	logsModalOpen = false;
	logsContainerId = "";
	logsContainerName = "";
}

// Обработчик перезагрузки контейнера через WebSocket
const handleRestartContainer = (containerId, containerName) => {
	if (!containerRestart) {
		console.warn("Container restart is disabled");
		return;
	}

	wsStore.connectRestartWebSocket(containerId, {
		onMessage: (data) => {
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
				alert(`Failed to restart container ${containerName}: ${data.message}`);
			}
		},
		onError: (error) => {
			console.error(`WebSocket error for restart ${containerName}:`, error);
			alert(`Failed to restart container ${containerName}: Connection error`);
		},
	});
};

onMount(() => {
	// Подключение WebSocket для контейнеров
	wsStore.connectContainersWebSocket({
		onOpen: () => {
			loading = false;
		},
		onMessage: (data) => {
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
		},
		onError: () => {
			loading = false;
		},
	});

	// Подключение WebSocket для метрик хоста
	wsStore.connectHostInfoWebSocket({
		onMessage: (data) => {
			hostinfo = data;
		},
	});

	// Подключение WebSocket для статистики контейнеров
	wsStore.connectStatsWebSocket({
		onMessage: (stats) => {
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
		},
	});

	checkMobileDevice();
	updateHeights();

	// Обновляем при изменении размеров окна
	const handleResize = () => {
		checkMobileDevice();
		updateHeights();
	};
	window.addEventListener("resize", handleResize);

	// Следим за изменением медиа-запроса
	const mediaQuery = window.matchMedia("(max-width: 768px)");
	const handleMediaChange = () => {
		checkMobileDevice();
		updateHeights();
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
		updateHeights();
	}, 100);
}

onDestroy(() => {
	wsStore.disconnectAll();
});
</script>

<div class="scrollable-content">
  <Header {appVersion} />

  <MetricsBar
    {hostinfo}
    {containersData}
    {isMobile}
    {metricsRowExpanded}
    {headerHeight}
    onToggle={handleToggleMetricsRow}
  />

  <FilterBar bind:filterText {headerHeight} {metricsBarHeight} />

  <GroupTagsBar
    {allGroups}
    bind:selectedGroup
    {isMobile}
    {groupTagsExpanded}
    {headerHeight}
    {metricsBarHeight}
    {filterBarHeight}
    onToggle={handleToggleGroupTags}
  />

  <ContainerGroups
    {filteredGroups}
    {containerStats}
    {logsShow}
    {containerRestart}
    {loading}
    {totalFixedHeight}
    onOpenLogs={openLogsModal}
    onRestartContainer={handleRestartContainer}
  />
</div>

<LogsModal
  open={logsModalOpen}
  containerId={logsContainerId}
  containerName={logsContainerName}
  on:close={closeLogsModal}
/>

<style>
  :global(body) {
    font-family: Arial, sans-serif;
    background-color: #234255;
    color: #333;
    margin: 0;
    padding: 0;
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
</style>
