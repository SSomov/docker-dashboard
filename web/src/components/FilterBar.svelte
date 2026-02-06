<script>
export let filterText = "";
export let headerHeight = 0;
export let metricsBarHeight = 0;
</script>

<div class="filter-bar" style="top: {headerHeight + metricsBarHeight}px">
  <label for="container-filter">Filter:</label>
  <input
    id="container-filter"
    type="text"
    placeholder="Filter containers by name..."
    bind:value={filterText}
  />
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
</style>

