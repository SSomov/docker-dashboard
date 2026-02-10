<script>
export let filterText = "";
export let headerHeight = 0;
export let metricsBarHeight = 0;
export let viewMode = "tile";
export let onViewModeChange = (mode) => {};

const handleViewModeChange = (mode) => {
  onViewModeChange(mode);
};
</script>

<div class="filter-bar" style="top: {headerHeight + metricsBarHeight}px">
  <label for="container-filter">Filter:</label>
  <input
    id="container-filter"
    type="text"
    placeholder="Filter containers by name..."
    bind:value={filterText}
  />
  <div class="view-mode-buttons">
    <button
      class="view-mode-button"
      class:active={viewMode === "tile"}
      on:click={() => handleViewModeChange("tile")}
      aria-label="Tile view"
      title="Tile view"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect x="3" y="3" width="7" height="7" />
        <rect x="14" y="3" width="7" height="7" />
        <rect x="3" y="14" width="7" height="7" />
        <rect x="14" y="14" width="7" height="7" />
      </svg>
    </button>
    <button
      class="view-mode-button"
      class:active={viewMode === "table"}
      on:click={() => handleViewModeChange("table")}
      aria-label="Table view"
      title="Table view"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        xmlns="http://www.w3.org/2000/svg"
      >
        <line x1="8" y1="6" x2="21" y2="6" />
        <line x1="8" y1="12" x2="21" y2="12" />
        <line x1="8" y1="18" x2="21" y2="18" />
        <line x1="3" y1="6" x2="3.01" y2="6" />
        <line x1="3" y1="12" x2="3.01" y2="12" />
        <line x1="3" y1="18" x2="3.01" y2="18" />
      </svg>
    </button>
  </div>
</div>

<style>
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

  .view-mode-buttons {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .view-mode-button {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 4px;
    padding: 0.5rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    color: rgba(255, 255, 255, 0.7);
    transition: all 0.2s;
    width: 36px;
    height: 36px;
  }

  .view-mode-button:hover {
    background: rgba(255, 255, 255, 0.15);
    color: rgba(255, 255, 255, 0.9);
    border-color: rgba(255, 255, 255, 0.5);
  }

  .view-mode-button.active {
    background: #4caf50;
    border-color: #4caf50;
    color: white;
  }

  .view-mode-button svg {
    width: 20px;
    height: 20px;
  }
</style>

