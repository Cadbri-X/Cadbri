<script lang="ts">
  import type { SearchResponse } from '../../types/search';

  let { response, page = 1, onPageChange } = $props();
</script>

<div class="w-full max-w-5xl mx-auto px-4 sm:px-6 py-6">
  {#if response.results && response.results.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
      {#each response.results as item}
        <article class="p-5 rounded-2xl bg-zinc-50/80 dark:bg-zinc-900/80 border border-zinc-200/80 dark:border-zinc-800 transition-all duration-200 hover:shadow-md flex flex-col justify-between">
          <div>
            <div class="flex items-center gap-2 mb-2 text-blue-600 dark:text-blue-400">
              <svg class="w-4 h-4 shrink-0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
              <span class="text-xs font-semibold uppercase tracking-wider">Location / Map</span>
            </div>

            <h3 class="text-base sm:text-lg font-bold text-zinc-900 dark:text-white leading-snug mb-1">
              <a href={item.url} target="_blank" rel="noopener noreferrer" class="hover:text-blue-600 dark:hover:text-blue-400">
                {item.title}
              </a>
            </h3>

            {#if item.content}
              <p class="text-sm text-zinc-600 dark:text-zinc-300 leading-relaxed line-clamp-3 mt-1.5">
                {item.content}
              </p>
            {/if}
          </div>

          <div class="mt-4 pt-3 border-t border-zinc-200/60 dark:border-zinc-800/60 flex items-center justify-between">
            <span class="text-xs text-zinc-500 dark:text-zinc-400 font-mono">
              {item.engines ? item.engines.join(', ') : 'OpenStreetMap'}
            </span>
            <a 
              href={item.url} 
              target="_blank" 
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium transition-colors"
            >
              <span>Open in Map</span>
              <svg class="w-3 h-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            </a>
          </div>
        </article>
      {/each}
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-center gap-3 pt-10 pb-12">
      {#if page > 1}
        <button
          type="button"
          onclick={() => onPageChange(page - 1)}
          class="px-4 py-2 rounded-full border border-blue-600 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-950/40 text-sm font-semibold transition-colors"
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
        class="px-4 py-2 rounded-full bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold transition-colors shadow-xs"
      >
        Next →
      </button>
    </div>
  {:else}
    <div class="py-20 text-center">
      <div class="w-12 h-12 mx-auto rounded-full bg-zinc-100 dark:bg-zinc-900 flex items-center justify-center text-zinc-400 mb-3">
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
      </div>
      <h3 class="text-base font-medium text-zinc-800 dark:text-zinc-200">No map locations found</h3>
      <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Try searching for a city, landmark, or address.</p>
    </div>
  {/if}
</div>
