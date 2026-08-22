<script lang="ts">
  import type { SearchResponse, SearchResult } from '../../types/search';

  let { response, page = 1, onPageChange } = $props();

  let selectedImage = $state<SearchResult | null>(null);

  function getHostname(url: string) {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }

  function handleImageError(e: Event) {
    const target = e.currentTarget as HTMLElement | null;
    target?.parentElement?.classList.add('hidden');
  }
</script>

<div class="w-full max-w-7xl mx-auto px-4 sm:px-6 py-6">
  {#if response.results && response.results.length > 0}
    <!-- Image Gallery Grid -->
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3 sm:gap-4">
      {#each response.results as item}
        {@const imgSrc = item.img_src || item.thumbnail || item.url}
        <button
          type="button"
          onclick={() => selectedImage = item}
          class="group relative flex flex-col overflow-hidden rounded-xl bg-zinc-100 dark:bg-zinc-900 border border-zinc-200/80 dark:border-zinc-800 text-left transition-all duration-200 hover:shadow-md hover:scale-[1.02] focus:outline-none cursor-pointer"
        >
          <!-- Image Thumbnail Container -->
          <div class="relative aspect-4/3 w-full overflow-hidden bg-zinc-200 dark:bg-zinc-800">
            <img 
              src={imgSrc} 
              alt={item.title} 
              class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
              loading="lazy"
              onerror={handleImageError}
            />
            <!-- Gradient Overlay on Hover -->
            <div class="absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex items-end p-2.5">
              <span class="text-xs text-white font-medium truncate">
                {getHostname(item.url)}
              </span>
            </div>
          </div>

          <!-- Bottom Caption -->
          <div class="p-2.5">
            <h4 class="text-xs font-medium text-zinc-800 dark:text-zinc-200 line-clamp-1 group-hover:text-blue-600 dark:group-hover:text-blue-400">
              {item.title}
            </h4>
            <span class="text-[10px] text-zinc-400 dark:text-zinc-500 mt-0.5 block truncate">
              {getHostname(item.url)}
            </span>
          </div>
        </button>
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
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="3" rx="2" ry="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/></svg>
      </div>
      <h3 class="text-base font-medium text-zinc-800 dark:text-zinc-200">No images found</h3>
      <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Try another search term.</p>
    </div>
  {/if}

  <!-- Lightbox Modal -->
  {#if selectedImage}
    {@const modalImgSrc = selectedImage.img_src || selectedImage.thumbnail || selectedImage.url}
    <div 
      class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6 animate-fade-in cursor-pointer"
      role="button"
      tabindex="0"
      aria-label="Close image modal"
      onclick={() => selectedImage = null}
      onkeydown={(e) => e.key === 'Escape' && (selectedImage = null)}
    >
      <div 
        class="relative max-w-4xl w-full bg-zinc-900 border border-zinc-800 rounded-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh] cursor-default"
        role="dialog"
        aria-modal="true"
        tabindex="-1"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => e.stopPropagation()}
      >
        <!-- Modal Top Bar -->
        <div class="flex items-center justify-between px-5 py-3 border-b border-zinc-800 bg-zinc-950">
          <div class="truncate mr-4">
            <h3 class="text-sm font-semibold text-white truncate">{selectedImage.title}</h3>
            <span class="text-xs text-zinc-400 truncate block">{getHostname(selectedImage.url)}</span>
          </div>
          <button
            type="button"
            onclick={() => selectedImage = null}
            aria-label="Close image modal"
            class="p-1.5 rounded-full text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors cursor-pointer"
          >
            <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <!-- Main Image Container -->
        <div class="flex-1 overflow-auto bg-black flex items-center justify-center p-4">
          <img 
            src={modalImgSrc} 
            alt={selectedImage.title} 
            class="max-h-[65vh] w-auto max-w-full object-contain rounded-lg"
          />
        </div>

        <!-- Modal Bottom Actions -->
        <div class="px-5 py-3 border-t border-zinc-800 bg-zinc-950 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-xs text-zinc-400 font-mono truncate max-w-xs">{selectedImage.url}</span>
          </div>
          <div class="flex items-center gap-2">
            <a 
              href={selectedImage.url} 
              target="_blank" 
              rel="noopener noreferrer"
              class="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-full text-xs font-semibold transition-colors flex items-center gap-1.5 cursor-pointer"
            >
              <span>Visit Page</span>
              <svg class="w-3 h-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            </a>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
