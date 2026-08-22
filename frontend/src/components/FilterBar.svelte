<script lang="ts">
  import type { SearchFilters } from '../types/search';

  let { filters = $bindable(), onFilterChange } = $props();

  let showFilters = $state(false);

  function handleTimeRangeChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    filters.timeRange = target.value as SearchFilters['timeRange'];
    onFilterChange();
  }

  function handleSafeSearchChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    filters.safeSearch = Number(target.value);
    onFilterChange();
  }

  function handleLanguageChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    filters.language = target.value;
    onFilterChange();
  }
</script>

<div class="flex items-center gap-2 shrink-0 py-0.5 text-xs">
  <button
    type="button"
    onclick={() => showFilters = !showFilters}
    class="flex items-center gap-1.5 px-3 py-1 rounded-full border border-zinc-200 dark:border-zinc-800 text-zinc-600 dark:text-zinc-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50/50 dark:hover:bg-blue-950/30 transition-colors cursor-pointer {showFilters ? 'bg-blue-50 border-blue-500 text-blue-600 dark:bg-blue-950/50 dark:border-blue-500 dark:text-blue-400 font-medium' : ''}"
  >
    <svg class="w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <line x1="4" x2="20" y1="21" y2="21"/>
      <line x1="4" x2="20" y1="3" y2="3"/>
      <line x1="4" x2="20" y1="12" y2="12"/>
      <line x1="10" x2="10" y1="9" y2="15"/>
      <line x1="16" x2="16" y1="18" y2="24"/>
      <line x1="8" x2="8" y1="0" y2="6"/>
    </svg>
    <span>Tools</span>
  </button>

  {#if showFilters}
    <div class="flex flex-wrap items-center gap-2 animate-fade-in pl-2 border-l border-zinc-200 dark:border-zinc-800">
      <!-- Time Filter -->
      <select
        value={filters.timeRange}
        onchange={handleTimeRangeChange}
        class="bg-transparent border border-zinc-200 dark:border-zinc-800 rounded-full px-2.5 py-1 text-zinc-700 dark:text-zinc-300 outline-none focus:border-blue-500 cursor-pointer"
      >
        <option value="">Anytime</option>
        <option value="day">Past 24 hours</option>
        <option value="week">Past week</option>
        <option value="month">Past month</option>
        <option value="year">Past year</option>
      </select>

      <!-- SafeSearch -->
      <select
        value={filters.safeSearch}
        onchange={handleSafeSearchChange}
        class="bg-transparent border border-zinc-200 dark:border-zinc-800 rounded-full px-2.5 py-1 text-zinc-700 dark:text-zinc-300 outline-none focus:border-blue-500 cursor-pointer"
      >
        <option value={0}>SafeSearch: Off</option>
        <option value={1}>SafeSearch: Moderate</option>
        <option value={2}>SafeSearch: Strict</option>
      </select>

      <!-- Language -->
      <select
        value={filters.language}
        onchange={handleLanguageChange}
        class="bg-transparent border border-zinc-200 dark:border-zinc-800 rounded-full px-2.5 py-1 text-zinc-700 dark:text-zinc-300 outline-none focus:border-blue-500 cursor-pointer"
      >
        <option value="all">All Languages</option>
        <option value="en">English</option>
        <option value="es">Español</option>
        <option value="fr">Français</option>
        <option value="de">Deutsch</option>
        <option value="zh">中文</option>
        <option value="ja">日本語</option>
      </select>
    </div>
  {/if}
</div>
