<script lang="ts">
  import Logo from './Logo.svelte';
  import SearchBar from './SearchBar.svelte';
  import CategoryTabs from './CategoryTabs.svelte';
  import FilterBar from './FilterBar.svelte';
  import type { CategoryId, SearchFilters } from '../types/search';

  let {
    query = $bindable(''),
    activeCategory = $bindable('general'),
    filters = $bindable(),
    isDarkMode = $bindable(false),
    onSearch,
    onNavigateHome
  } = $props();

  function toggleDarkMode() {
    isDarkMode = !isDarkMode;
    if (isDarkMode) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('cadbri_theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('cadbri_theme', 'light');
    }
  }
</script>

<header class="sticky top-0 z-40 w-full bg-white/95 dark:bg-zinc-950/95 backdrop-blur-md border-b border-zinc-200 dark:border-zinc-800 transition-colors">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 pt-3 pb-2.5">
    <!-- Top Row: Logo, SearchBar, Controls -->
    <div class="flex items-center gap-4 sm:gap-6">
      <!-- Minimalist Logo with fixed width for exact alignment -->
      <button 
        type="button" 
        onclick={onNavigateHome}
        class="flex items-center gap-2.5 group text-left shrink-0 w-auto sm:w-28 focus:outline-none cursor-pointer"
      >
        <Logo size="sm" />
        <span class="text-lg font-bold tracking-tight text-zinc-900 dark:text-white hidden sm:inline">
          Cadbri
        </span>
      </button>

      <!-- Centered Compact SearchBar -->
      <div class="flex-1 max-w-2xl">
        <SearchBar bind:value={query} {onSearch} size="normal" />
      </div>

      <!-- Right Controls: Theme Toggle -->
      <div class="flex items-center justify-end gap-2 shrink-0 w-auto sm:w-10">
        <button
          type="button"
          onclick={toggleDarkMode}
          aria-label="Toggle theme"
          class="p-2 rounded-full text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors cursor-pointer"
        >
          {#if isDarkMode}
            <!-- Sun Icon -->
            <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>
            </svg>
          {:else}
            <!-- Moon Icon -->
            <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 3a6 6 0 0 0 9 9 9 0 1 1-9-9Z"/>
            </svg>
          {/if}
        </button>
      </div>
    </div>

    <!-- Bottom Row: Category Tabs & Filter Tools aligned directly under SearchBar in single row -->
    <div class="flex items-center gap-4 sm:gap-6 mt-2 pt-1 sm:mt-1 sm:pt-0">
      <!-- Spacer matching Logo width on sm+ screens -->
      <div class="hidden sm:block sm:w-28 shrink-0"></div>

      <!-- Category Tabs & Tools Container aligned directly under SearchBar in single row -->
      <div class="flex-1 max-w-2xl flex items-center gap-2 overflow-x-auto no-scrollbar py-0.5">
        <CategoryTabs 
          activeTab={activeCategory} 
          onSelectTab={(cat) => {
            activeCategory = cat;
            if (query.trim()) onSearch(query.trim());
          }} 
        />
        <FilterBar bind:filters onFilterChange={() => query.trim() && onSearch(query.trim())} />
      </div>

      <!-- Spacer matching right controls -->
      <div class="hidden sm:block sm:w-10 shrink-0"></div>
    </div>
  </div>
</header>
