<script>
import { createEventDispatcher, onDestroy, onMount } from "svelte";

export let open = false;
export let containerId = "";
export let containerName = "";

const dispatch = createEventDispatcher();

let logs = [];
let wsLogs = null;
let logsContainerRef = null;
let autoScrollEnabled = true;

function isScrolledToBottom(element) {
	if (!element) return true;
	const threshold = 10; // Небольшой порог для учета погрешности
	return (
		element.scrollHeight - element.scrollTop - element.clientHeight < threshold
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
	if (event.key === "Escape" && open) {
		closeModal();
	}
}

function handleModalContentClick(event) {
	event.stopPropagation();
}

function closeModal() {
	if (wsLogs) {
		wsLogs.close();
		wsLogs = null;
	}
	logs = [];
	dispatch("close");
}

function connectLogsWebSocket() {
	if (!containerId) return;

	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	const wsUrl = `${protocol}//${window.location.host}${window.location.pathname}ws/containers/${containerId}/logs`;

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

$: if (open && containerId && !wsLogs) {
	logs = [];
	autoScrollEnabled = true; // Сбрасываем в true при открытии
	connectLogsWebSocket();
}

$: if (!open && wsLogs) {
	wsLogs.close();
	wsLogs = null;
	logs = [];
}

onMount(() => {
	window.addEventListener("keydown", handleEscapeKey);
});

onDestroy(() => {
	if (wsLogs) {
		wsLogs.close();
	}
	window.removeEventListener("keydown", handleEscapeKey);
});
</script>

{#if open}
  <div
    class="modal-overlay"
    role="button"
    tabindex="0"
    on:click={closeModal}
    on:keydown={(e) => e.key === "Enter" && closeModal()}
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
        <h2 id="modal-title">Logs: {containerName}</h2>
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
            on:click={closeModal}
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
</style>

