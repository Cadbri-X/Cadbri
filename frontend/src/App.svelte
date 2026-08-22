<script lang="ts">
  import { onMount } from 'svelte';
  import Header from './components/Header.svelte';
  import HomePage from './components/HomePage.svelte';
  import WebResults from './components/views/WebResults.svelte';
  import ImageResults from './components/views/ImageResults.svelte';
  import VideoResults from './components/views/VideoResults.svelte';
  import NewsResults from './components/views/NewsResults.svelte';
  import MapResults from './components/views/MapResults.svelte';
  import { searchCadbri } from './api/client';
  import type { SearchResponse, CategoryId, SearchFilters } from './types/search';

  let query = $state('');
  let activeCategory = $state<CategoryId>('general');
  let page = $state(1);
  let searchResponse = $state<SearchResponse | null>(null);
  let searchDuration = $state('0.00');
  let isLoading = $state(false);
  let errorMessage = $state<string | null>(null);
  let isDarkMode = $state(false);

  let filters = $state<SearchFilters>({
    timeRange: '',
    safeSearch: 0,
    language: 'all'
  });

  let currentAbortController: AbortController | null = null;

  async function executeSearch(newQuery: string, newCategory: CategoryId = activeCategory, newPage: number = 1) {
    const trimmed = newQuery.trim();
    if (!trimmed) {
      query = '';
      searchResponse = null;
      searchDuration = '0.00';
      window.history.pushState({}, '', '/');
      return;
    }

    query = trimmed;
    activeCategory = newCategory;
    page = newPage;
    isLoading = true;
    errorMessage = null;

    // Update Browser URL params
    const params = new URLSearchParams({ q: query });
    if (activeCategory !== 'general') params.set('category', activeCategory);
    if (page > 1) params.set('page', page.toString());
    if (filters.timeRange) params.set('time_range', filters.timeRange);
    if (filters.safeSearch > 0) params.set('safesearch', filters.safeSearch.toString());
    if (filters.language !== 'all') params.set('language', filters.language);

    window.history.pushState({}, '', `/?${params.toString()}`);
    document.title = `${query} - Cadbri Search`;

    if (currentAbortController) {
      currentAbortController.abort();
    }
    currentAbortController = new AbortController();

    const startTime = performance.now();
    try {
      const resp = await searchCadbri(
        query,
        activeCategory,
        page,
        filters,
        currentAbortController.signal
      );
      searchDuration = ((performance.now() - startTime) / 1000).toFixed(2);
      searchResponse = resp;
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        errorMessage = err.message || 'Failed to retrieve search results.';
      }
    } finally {
      isLoading = false;
    }
  }

  function handleHomeSearch(q: string, cat?: CategoryId) {
    executeSearch(q, cat || 'general', 1);
  }

  function handleHeaderSearch(q: string) {
    executeSearch(q, activeCategory, 1);
  }

  function handlePageChange(p: number) {
    executeSearch(query, activeCategory, p);
  }

  function handleSuggestionSearch(sug: string) {
    executeSearch(sug, activeCategory, 1);
  }

  function handleNavigateHome() {
    query = '';
    searchResponse = null;
    errorMessage = null;
    window.history.pushState({}, '', '/');
    document.title = 'Cadbri Search';
  }

  function handlePopState() {
    const params = new URLSearchParams(window.location.search);
    const q = params.get('q');
    if (q) {
      const cat = (params.get('category') as CategoryId) || 'general';
      const p = parseInt(params.get('page') || '1', 10);
      query = q;
      activeCategory = cat;
      page = p;
      executeSearch(q, cat, p);
    } else {
      query = '';
      searchResponse = null;
    }
  }

  onMount(() => {
    // 1. Initialize Theme
    const savedTheme = localStorage.getItem('cadbri_theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
      isDarkMode = true;
      document.documentElement.classList.add('dark');
    } else {
      isDarkMode = false;
      document.documentElement.classList.remove('dark');
    }

    // 2. Read initial URL
    const params = new URLSearchParams(window.location.search);
    const initialQ = params.get('q');
    if (initialQ) {
      const initialCat = (params.get('category') as CategoryId) || 'general';
      const initialPage = parseInt(params.get('page') || '1', 10);
      executeSearch(initialQ, initialCat, initialPage);
    }

    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  });
</script>

<div class="min-h-screen flex flex-col bg-white dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 selection:bg-blue-500 selection:text-white transition-colors duration-150">
  {#if !query && !searchResponse && !isLoading}
    <!-- Home Landing View -->
    <HomePage 
      onSearch={handleHomeSearch} 
      bind:isDarkMode 
    />
  {:else}
    <!-- Search Results View Header -->
    <Header
      bind:query
      bind:activeCategory
      bind:filters
      bind:isDarkMode
      onSearch={handleHeaderSearch}
      onNavigateHome={handleNavigateHome}
    />

    <!-- Main Results Body -->
    <main class="flex-1">
      {#if isLoading}
        <!-- Minimalist Loading Skeleton -->
        <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8">
          <div class="space-y-6 max-w-2xl">
            <div class="h-4 w-40 bg-zinc-100 dark:bg-zinc-900 rounded-full animate-pulse"></div>
            {#each Array(5) as _}
              <div class="space-y-2.5 p-4 rounded-2xl bg-zinc-50/50 dark:bg-zinc-900/40 border border-zinc-100 dark:border-zinc-800/50 animate-pulse">
                <div class="flex items-center gap-2">
                  <div class="w-4 h-4 rounded-full bg-zinc-200 dark:bg-zinc-800"></div>
                  <div class="h-3 w-28 bg-zinc-200 dark:bg-zinc-800 rounded"></div>
                </div>
                <div class="h-5 w-3/4 bg-zinc-200 dark:bg-zinc-800 rounded-md"></div>
                <div class="h-3 w-full bg-zinc-200/80 dark:bg-zinc-800/80 rounded"></div>
                <div class="h-3 w-5/6 bg-zinc-200/80 dark:bg-zinc-800/80 rounded"></div>
              </div>
            {/each}
          </div>
        </div>
      {:else if errorMessage}
        <!-- Error Banner -->
        <div class="max-w-xl mx-auto px-4 py-20 text-center">
          <div class="w-12 h-12 mx-auto rounded-full bg-red-50 dark:bg-red-950/50 text-red-600 dark:text-red-400 flex items-center justify-center mb-3">
            <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          </div>
          <h3 class="text-base font-semibold text-zinc-900 dark:text-white">Unable to fetch results</h3>
          <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">{errorMessage}</p>
          <button
            type="button"
            onclick={() => executeSearch(query, activeCategory, page)}
            class="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-full shadow-xs cursor-pointer"
          >
            Try Again
          </button>
        </div>
      {:else if searchResponse}
        <!-- Active Category View Dispatcher -->
        {#if activeCategory === 'images'}
          <ImageResults 
            response={searchResponse} 
            {page} 
            onPageChange={handlePageChange} 
          />
        {:else if activeCategory === 'videos'}
          <VideoResults 
            response={searchResponse} 
            {page} 
            onPageChange={handlePageChange} 
          />
        {:else if activeCategory === 'news'}
          <NewsResults 
            response={searchResponse} 
            {page} 
            onPageChange={handlePageChange} 
          />
        {:else if activeCategory === 'map'}
          <MapResults 
            response={searchResponse} 
            {page} 
            onPageChange={handlePageChange} 
          />
        {:else}
          <!-- Web, IT, Science, General -->
          <WebResults 
            response={searchResponse} 
            {page} 
            {searchDuration}
            onPageChange={handlePageChange}
            onSearchSuggestion={handleSuggestionSearch}
          />
        {/if}
      {/if}
    </main>

    <!-- Footer -->
    <footer class="border-t border-zinc-200 dark:border-zinc-850 py-4 px-6 text-center text-xs text-zinc-400 dark:text-zinc-500">
      <div class="max-w-7xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-2">
        <span>Cadbri Search &bull; Fast, Minimalist &amp; Privacy-First</span>
        <div class="flex items-center gap-4">
          <button type="button" onclick={handleNavigateHome} class="hover:text-zinc-600 dark:hover:text-zinc-300 cursor-pointer">Home</button>
          <span>&bull;</span>
          <span>Docker Native</span>
        </div>
      </div>
    </footer>
  {/if}
</div>
