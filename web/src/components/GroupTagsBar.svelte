<script>
import { createEventDispatcher } from "svelte";

export let allGroups = [];
export let selectedGroup = null;
export let isMobile = false;
export let groupTagsExpanded = false;
export let headerHeight = 0;
export let metricsBarHeight = 0;
export let filterBarHeight = 0;

const dispatch = createEventDispatcher();

function handleToggle() {
	dispatch("toggle");
}

function handleGroupClick(groupName) {
	if (selectedGroup === groupName) {
		selectedGroup = null; // Снимаем фильтр при повторном клике
	} else {
		selectedGroup = groupName;
	}
}
</script>

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
          on:click={handleToggle}
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

<style>
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
    .group-tags-container.hidden-mobile {
      max-height: none !important;
      opacity: 1 !important;
    }
  }
</style>

