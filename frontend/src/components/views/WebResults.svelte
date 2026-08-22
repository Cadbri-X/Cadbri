<script lang="ts">
  import type { SearchResponse } from '../../types/search';

  let {
    response,
    page = 1,
    searchDuration = '0.00',
    onPageChange,
    onSearchSuggestion
  } = $props();

  function getFavicon(url: string) {
    try {
      const u = new URL(url);
      return `https://www.google.com/s2/favicons?domain=${u.hostname}&sz=64`;
    } catch {
      return '';
    }
  }

  function getHostname(url: string) {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }

  function formatPath(url: string) {
    try {
      const u = new URL(url);
      const parts = u.pathname.split('/').filter(Boolean);
      if (parts.length === 0) return '';
      return ' › ' + parts.slice(0, 3).join(' › ');
    } catch {
      return '';
    }
  }

  function handleFaviconError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    if (target) target.style.display = 'none';
  }

  function handleInfoboxImageError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    if (target && target.parentElement) {
      target.parentElement.style.display = 'none';
    }
  }

  let uniqueInfoboxes = $derived.by(() => {
    if (!response.infoboxes) return [];
    const seen = new Set<string>();
    return response.infoboxes.filter((info) => {
      const key = (info.infobox || '').trim().toLowerCase();
      if (!key || seen.has(key)) return false;
      seen.add(key);
      return true;
    });
  });
</script>

<div class="w-full max-w-7xl mx-auto px-4 sm:px-6 py-4">
  <div class="flex flex-col lg:flex-row gap-4 sm:gap-6">
    <!-- Left offset spacer matching Header Logo like Google Search -->
    <div class="hidden sm:block sm:w-28 shrink-0"></div>

    <!-- Main Search Results Column (aligned directly under SearchBar) -->
    <div class="flex-1 max-w-2xl space-y-6">
      <!-- Result Stats with Search Latency Duration -->
      {#if response.number_of_results > 0}
        <div class="text-xs text-zinc-500 dark:text-zinc-400 font-medium">
          About {response.number_of_results.toLocaleString()} results found ({searchDuration} seconds)
        </div>
      {/if}

      <!-- Spelling Corrections -->
      {#if response.corrections && response.corrections.length > 0}
        <div class="p-3 bg-blue-50/50 dark:bg-blue-950/30 border border-blue-100 dark:border-blue-900/50 rounded-xl text-sm">
          <span class="text-zinc-600 dark:text-zinc-400">Did you mean:</span>
          {#each response.corrections as cor}
            <button
              type="button"
              onclick={() => onSearchSuggestion(cor)}
              class="ml-2 font-medium text-blue-600 dark:text-blue-400 hover:underline cursor-pointer"
            >
              {cor}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Organic Search Results List -->
      {#if response.results && response.results.length > 0}
        <div class="space-y-6">
          {#each response.results as item}
            <article class="group result-card-hover">
              <!-- Source & URL line with larger favicon -->
              <div class="flex items-center gap-2.5 mb-1.5">
                <div class="w-6 h-6 rounded-full bg-zinc-100 dark:bg-zinc-800 border border-zinc-200/80 dark:border-zinc-700/80 p-0.5 flex items-center justify-center shrink-0 shadow-2xs">
                  <img 
                    src={getFavicon(item.url)} 
                    alt="" 
                    class="w-full h-full object-contain rounded-full"
                    loading="lazy"
                    onerror={handleFaviconError}
                  />
                </div>
                <div class="flex items-baseline gap-1.5 min-w-0">
                  <span class="text-xs text-zinc-800 dark:text-zinc-200 font-semibold truncate max-w-xs">
                    {getHostname(item.url)}
                  </span>
                  <span class="text-[11px] text-zinc-400 dark:text-zinc-500 truncate max-w-sm hidden sm:inline">
                    {formatPath(item.url)}
                  </span>
                </div>
              </div>

              <!-- Title Link -->
              <h2 class="text-lg sm:text-xl font-medium tracking-tight mb-1.5 leading-snug">
                <a 
                  href={item.url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  class="text-blue-700 dark:text-blue-400 hover:underline group-hover:text-blue-800 dark:group-hover:text-blue-300 cursor-pointer"
                >
                  {item.title}
                </a>
              </h2>

              <!-- Snippet / Description -->
              <p class="text-sm text-zinc-600 dark:text-zinc-300 leading-relaxed line-clamp-3">
                {item.content}
              </p>

              <!-- Metadata & Engine Badges -->
              {#if item.engines && item.engines.length > 0}
                <div class="flex items-center gap-1.5 mt-2">
                  {#each item.engines as eng}
                    <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-medium bg-zinc-100 dark:bg-zinc-800/80 text-zinc-600 dark:text-zinc-400 border border-zinc-200/60 dark:border-zinc-700/60">
                      {eng}
                    </span>
                  {/each}
                  {#if item.publishedDate}
                    <span class="text-[11px] text-zinc-400 dark:text-zinc-500 ml-2">
                      {item.publishedDate}
                    </span>
                  {/if}
                </div>
              {/if}
            </article>
          {/each}
        </div>
      {:else}
        <div class="py-16 text-center">
          <div class="w-12 h-12 mx-auto rounded-full bg-zinc-100 dark:bg-zinc-900 flex items-center justify-center text-zinc-400 mb-3">
            <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
          </div>
          <h3 class="text-base font-medium text-zinc-800 dark:text-zinc-200">No results found</h3>
          <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Try different keywords or broader search filters.</p>
        </div>
      {/if}

      <!-- Related Suggestions -->
      {#if response.suggestions && response.suggestions.length > 0}
        <div class="pt-6 border-t border-zinc-200 dark:border-zinc-800">
          <h4 class="text-sm font-semibold text-zinc-800 dark:text-zinc-200 mb-3">
            Related Searches
          </h4>
          <div class="flex flex-wrap gap-2">
            {#each response.suggestions as sug}
              <button
                type="button"
                onclick={() => onSearchSuggestion(sug)}
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-zinc-100 dark:bg-zinc-850 hover:bg-zinc-200 dark:hover:bg-zinc-800 text-xs sm:text-sm text-zinc-700 dark:text-zinc-300 transition-colors cursor-pointer"
              >
                <svg class="w-3.5 h-3.5 text-zinc-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                <span>{sug}</span>
              </button>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Pagination -->
      <div class="flex items-center justify-center gap-3 pt-8 pb-12">
        {#if page > 1}
          <button
            type="button"
            onclick={() => onPageChange(page - 1)}
            class="px-4 py-2 rounded-full border border-blue-600 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-950/40 text-sm font-semibold transition-colors cursor-pointer"
          >
            ← Previous
          </button>
        {/if}
        <span class="text-sm text-zinc-500 dark:text-zinc-400 font-medium px-2">
          Page {page}
        </span>
        <button
          type="button"
          onclick={() => onPageChange(page + 1)}
          class="px-4 py-2 rounded-full bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold transition-colors shadow-xs cursor-pointer"
        >
          Next →
        </button>
      </div>
    </div>

    <!-- Sidebar Knowledge Infobox & Instant Answer Column (Right side) -->
    {#if (response.infoboxes && response.infoboxes.length > 0) || (response.answers && response.answers.length > 0)}
      <aside class="w-full lg:w-80 xl:w-96 shrink-0 lg:ml-6 space-y-6">
        <!-- 1. Rich Entity Infobox Cards -->
        {#if uniqueInfoboxes.length > 0}
          {#each uniqueInfoboxes as info}
            <div class="p-5 bg-zinc-50/80 dark:bg-zinc-900/80 border border-zinc-200 dark:border-zinc-800 rounded-2xl sticky top-28 shadow-xs animate-fade-in">
              {#if info.img_src}
                <div class="mb-4 overflow-hidden rounded-xl bg-zinc-100 dark:bg-zinc-800">
                  <img 
                    src={info.img_src} 
                    alt={info.infobox} 
                    class="w-full max-h-48 object-cover" 
                    loading="lazy"
                    referrerpolicy="no-referrer"
                    onerror={handleInfoboxImageError}
                  />
                </div>
              {/if}

              <h3 class="text-lg font-bold text-zinc-900 dark:text-white mb-2">
                {info.infobox}
              </h3>

              {#if info.content}
                <p class="text-sm text-zinc-600 dark:text-zinc-300 leading-relaxed mb-4">
                  {info.content}
                </p>
              {/if}

              {#if info.attributes && info.attributes.length > 0}
                <div class="space-y-1.5 py-3 border-t border-zinc-200/60 dark:border-zinc-800/60 text-xs">
                  {#each info.attributes as attr}
                    <div class="flex justify-between py-1 border-b border-zinc-100 dark:border-zinc-800/40">
                      <span class="text-zinc-500 dark:text-zinc-400 font-medium">{attr.label}</span>
                      <span class="text-zinc-900 dark:text-zinc-200 font-semibold">{attr.value}</span>
                    </div>
                  {/each}
                </div>
              {/if}

              {#if info.urls && info.urls.length > 0}
                <div class="pt-2">
                  {#each info.urls as u}
                    <a 
                      href={u.url} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      class="inline-flex items-center gap-1 text-xs font-medium text-blue-600 dark:text-blue-400 hover:underline cursor-pointer"
                    >
                      <span>{u.title || 'Learn more'}</span>
                      <svg class="w-3 h-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                    </a>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
        {:else if response.answers && response.answers.length > 0}
          <!-- 2. Instant Answers (when no infobox exists) -->
          {#each response.answers as ans}
            <div class="p-5 bg-zinc-50/80 dark:bg-zinc-900/80 border border-zinc-200 dark:border-zinc-800 rounded-2xl sticky top-28 shadow-xs animate-fade-in">
              <div class="text-xs font-semibold text-blue-600 dark:text-blue-400 uppercase tracking-wider mb-1.5">
                Instant Answer
              </div>
              <div class="text-base font-bold text-zinc-900 dark:text-white mb-2 leading-snug">
                {ans.answer}
              </div>
              {#if ans.url}
                <a href={ans.url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 text-xs font-medium text-blue-600 dark:text-blue-400 hover:underline mt-2 cursor-pointer">
                  <span>Source: {getHostname(ans.url)}</span>
                  <svg class="w-3 h-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                </a>
              {/if}
            </div>
          {/each}
        {/if}
      </aside>
    {/if}
  </div>
</div>
