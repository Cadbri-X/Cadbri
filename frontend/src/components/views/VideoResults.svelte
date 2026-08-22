<script lang="ts">
  import type { SearchResponse, SearchResult } from '../../types/search';

  let { response, page = 1, onPageChange } = $props();

  let playingVideo = $state<SearchResult | null>(null);

  function getHostname(url: string) {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }

  function handleVideoThumbError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    if (target) target.style.display = 'none';
  }

  function formatLength(len?: number): string {
    if (!len || len <= 0) return '';
    const mins = Math.floor(len / 60);
    const secs = len % 60;
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  }
</script>

<div class="w-full max-w-7xl mx-auto px-4 sm:px-6 py-6">
  {#if response.results && response.results.length > 0}
    <!-- Video Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
      {#each response.results as item}
        {@const thumb = item.thumbnail || item.img_src}
        <article class="group flex flex-col rounded-2xl overflow-hidden bg-zinc-50/70 dark:bg-zinc-900/60 border border-zinc-200/80 dark:border-zinc-800 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
          <!-- Thumbnail with Play Button -->
          <div class="relative aspect-video w-full overflow-hidden bg-zinc-200 dark:bg-zinc-800">
            {#if thumb}
              <img 
                src={thumb} 
                alt={item.title} 
                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                loading="lazy"
                onerror={handleVideoThumbError}
              />
            {/if}

            <!-- Duration Badge -->
            {#if item.length}
              <span class="absolute bottom-2 right-2 px-1.5 py-0.5 rounded-md bg-black/80 text-white text-[11px] font-mono font-medium">
                {formatLength(item.length)}
              </span>
            {/if}

            <!-- Play Overlay Button -->
            <button
              type="button"
              onclick={() => playingVideo = item}
              aria-label="Play video"
              class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center cursor-pointer"
            >
              <div class="w-12 h-12 rounded-full bg-blue-600/90 text-white flex items-center justify-center shadow-lg transform transition-transform group-hover:scale-110">
                <svg class="w-5 h-5 ml-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              </div>
            </button>
          </div>

          <!-- Video Details -->
          <div class="p-3.5 flex-1 flex flex-col justify-between">
            <div>
              <h3 class="text-sm font-semibold text-zinc-900 dark:text-zinc-100 line-clamp-2 group-hover:text-blue-600 dark:group-hover:text-blue-400 leading-snug">
                <a href={item.url} target="_blank" rel="noopener noreferrer" class="cursor-pointer">
                  {item.title}
                </a>
              </h3>
              {#if item.content}
                <p class="text-xs text-zinc-500 dark:text-zinc-400 mt-1 line-clamp-2 leading-relaxed">
                  {item.content}
                </p>
              {/if}
            </div>

            <div class="flex items-center justify-between mt-3 pt-2.5 border-t border-zinc-200/50 dark:border-zinc-800/50 text-[11px] text-zinc-500 dark:text-zinc-400">
              <span class="font-medium truncate max-w-32.5">
                {item.author || getHostname(item.url)}
              </span>
              <span>
                {item.views || item.publishedDate || ''}
              </span>
            </div>
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
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m22 8-6 4 6 4V8Z"/><rect width="14" height="12" x="2" y="6" rx="2" ry="2"/></svg>
      </div>
      <h3 class="text-base font-medium text-zinc-800 dark:text-zinc-200">No videos found</h3>
      <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Try another search query.</p>
    </div>
  {/if}

  <!-- Video Playback Modal -->
  {#if playingVideo}
    <div 
      class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6 animate-fade-in cursor-pointer"
      role="button"
      tabindex="0"
      aria-label="Close video player"
      onclick={() => playingVideo = null}
      onkeydown={(e) => e.key === 'Escape' && (playingVideo = null)}
    >
      <div 
        class="relative max-w-4xl w-full bg-zinc-950 border border-zinc-800 rounded-2xl overflow-hidden shadow-2xl flex flex-col cursor-default"
        role="dialog"
        aria-modal="true"
        tabindex="-1"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => e.stopPropagation()}
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-5 py-3 border-b border-zinc-800">
          <h3 class="text-sm font-semibold text-white truncate pr-4">{playingVideo.title}</h3>
          <button
            type="button"
            onclick={() => playingVideo = null}
            aria-label="Close video modal"
            class="p-1.5 rounded-full text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors cursor-pointer"
          >
            <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <!-- Player Frame -->
        <div class="aspect-video w-full bg-black">
          {#if playingVideo.iframe_src}
            <iframe 
              src={playingVideo.iframe_src} 
              title={playingVideo.title}
              class="w-full h-full border-0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
            ></iframe>
          {:else if playingVideo.url.includes('youtube.com/watch?v=')}
            {@const videoId = new URL(playingVideo.url).searchParams.get('v')}
            <iframe 
              src="https://www.youtube-nocookie.com/embed/{videoId}?autoplay=1" 
              title={playingVideo.title}
              class="w-full h-full border-0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
            ></iframe>
          {:else}
            <div class="w-full h-full flex flex-col items-center justify-center text-zinc-400 p-6 text-center">
              <p class="text-sm mb-4">Direct embedded player not supported for this source.</p>
              <a 
                href={playingVideo.url} 
                target="_blank" 
                rel="noopener noreferrer"
                class="px-5 py-2 rounded-full bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold flex items-center gap-2 cursor-pointer"
              >
                <span>Watch on Source</span>
                <svg class="w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
              </a>
            </div>
          {/if}
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-zinc-800 flex items-center justify-between text-xs text-zinc-400">
          <span class="font-medium truncate max-w-sm">{playingVideo.author || getHostname(playingVideo.url)}</span>
          <a 
            href={playingVideo.url} 
            target="_blank" 
            rel="noopener noreferrer"
            class="text-blue-400 hover:underline flex items-center gap-1 cursor-pointer"
          >
            <span>Open Link</span>
            <svg class="w-3 h-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
          </a>
        </div>
      </div>
    </div>
  {/if}
</div>
