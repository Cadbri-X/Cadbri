<script lang="ts">
  import type { SearchResponse } from '../../types/search';

  let { response, page = 1, onPageChange } = $props();

  function getHostname(url: string) {
    try {
      return new URL(url).hostname.replace(/^www\./, '');
    } catch {
      return url;
    }
  }

  function getFavicon(url: string) {
    try {
      const u = new URL(url);
      return `https://www.google.com/s2/favicons?domain=${u.hostname}&sz=64`;
    } catch {
      return '';
    }
  }

  function handleFaviconError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    if (target) target.style.display = 'none';
  }

  function handleThumbError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    target?.parentElement?.remove();
  }
</script>

<div class="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6">
  {#if response.results && response.results.length > 0}
    <div class="space-y-4 sm:space-y-6">
      {#each response.results as item}
        {@const thumb = item.thumbnail || item.img_src}
        <article class="p-4 sm:p-5 rounded-2xl bg-zinc-50/80 dark:bg-zinc-900/80 border border-zinc-200/80 dark:border-zinc-800 transition-all duration-200 hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-700 flex flex-col sm:flex-row gap-4 items-start justify-between">
          <div class="flex-1 min-w-0">
            <!-- News Outlet & Date with larger favicon -->
            <div class="flex items-center gap-2.5 mb-2">
              <div class="w-6 h-6 rounded-full bg-zinc-100 dark:bg-zinc-800 border border-zinc-200/80 dark:border-zinc-700/80 p-0.5 flex items-center justify-center shrink-0 shadow-2xs">
                <img 
                  src={getFavicon(item.url)} 
                  alt="" 
                  class="w-full h-full object-contain rounded-full" 
                  loading="lazy"
                  onerror={handleFaviconError}
                />
              </div>
              <span class="text-xs font-semibold text-zinc-900 dark:text-zinc-200 uppercase tracking-wider">
                {item.author || getHostname(item.url)}
              </span>
              {#if item.publishedDate || item.pubdate}
                <span class="text-xs text-zinc-400 dark:text-zinc-500">
                  • {item.publishedDate || item.pubdate}
                </span>
              {/if}
            </div>

            <!-- Headline -->
            <h3 class="text-base sm:text-lg font-semibold text-zinc-900 dark:text-white leading-snug mb-1.5">
              <a 
                href={item.url} 
                target="_blank" 
                rel="noopener noreferrer"
                class="hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
              >
                {item.title}
              </a>
            </h3>

            <!-- Article Excerpt -->
            {#if item.content}
              <p class="text-sm text-zinc-600 dark:text-zinc-300 line-clamp-2 leading-relaxed">
                {item.content}
              </p>
            {/if}

            <!-- Engines Badge -->
            {#if item.engines && item.engines.length > 0}
              <div class="mt-2.5 flex items-center gap-1.5">
                {#each item.engines as eng}
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-zinc-200/60 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400">
                    {eng}
                  </span>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Optional Article Image Thumbnail -->
          {#if thumb}
            <div class="w-full sm:w-32 h-24 sm:h-24 rounded-xl overflow-hidden bg-zinc-200 dark:bg-zinc-800 shrink-0">
              <img 
                src={thumb} 
                alt={item.title} 
                class="w-full h-full object-cover" 
                loading="lazy"
                onerror={handleThumbError}
              />
            </div>
          {/if}
        </article>
      {/each}
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-center gap-3 pt-10 pb-12">
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
  {:else}
    <div class="py-20 text-center">
      <div class="w-12 h-12 mx-auto rounded-full bg-zinc-100 dark:bg-zinc-900 flex items-center justify-center text-zinc-400 mb-3">
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2Zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2"/><path d="M18 14h-8"/><path d="M15 18h-5"/><path d="M10 6h8v4h-8V6Z"/></svg>
      </div>
      <h3 class="text-base font-medium text-zinc-800 dark:text-zinc-200">No news articles found</h3>
      <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Try another news query.</p>
    </div>
  {/if}
</div>
