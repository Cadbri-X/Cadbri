<script lang="ts">
  import { fetchSuggestions } from '../api/client';

  let {
    value = $bindable(''),
    onSearch,
    placeholder = 'Search with Cadbri or type a URL...',
    autoFocus = false,
    size = 'normal'
  } = $props();

  let suggestions = $state<string[]>([]);
  let isFocused = $state(false);
  let selectedIndex = $state(-1);
  let inputElement = $state<HTMLInputElement | null>(null);
  let debounceTimeout: number | undefined;

  function handleInput(e: Event) {
    const val = (e.target as HTMLInputElement).value;
    value = val;
    selectedIndex = -1;

    clearTimeout(debounceTimeout);
    if (!val.trim()) {
      suggestions = [];
      return;
    }

    debounceTimeout = window.setTimeout(async () => {
      try {
        suggestions = await fetchSuggestions(val);
      } catch {
        suggestions = [];
      }
    }, 150);
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (suggestions.length > 0) {
        selectedIndex = (selectedIndex + 1) % suggestions.length;
        value = suggestions[selectedIndex];
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (suggestions.length > 0) {
        selectedIndex = selectedIndex <= 0 ? suggestions.length - 1 : selectedIndex - 1;
        value = suggestions[selectedIndex];
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      suggestions = [];
      if (value.trim()) {
        onSearch(value.trim());
      }
    } else if (e.key === 'Escape') {
      suggestions = [];
      selectedIndex = -1;
      inputElement?.blur();
    }
  }

  function selectSuggestion(s: string) {
    value = s;
    suggestions = [];
    selectedIndex = -1;
    onSearch(s);
  }

  function clearSearch() {
    value = '';
    suggestions = [];
    selectedIndex = -1;
    inputElement?.focus();
  }

  $effect(() => {
    if (autoFocus && inputElement) {
      inputElement.focus();
    }
  });
</script>

<div class="relative w-full">
  <div 
    class="flex items-center w-full transition-all duration-200 rounded-full border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 shadow-sm hover:shadow-md focus-within:shadow-md focus-within:border-blue-500 dark:focus-within:border-blue-500 {size === 'large' ? 'px-5 py-3.5 text-base' : 'px-4 py-2 text-sm'}"
  >
    <!-- Search Icon -->
    <svg class="w-4 h-4 text-zinc-400 dark:text-zinc-500 shrink-0 mr-3" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="11" cy="11" r="8"></circle>
      <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
    </svg>

    <!-- Input Field -->
    <input
      bind:this={inputElement}
      type="search"
      {value}
      oninput={handleInput}
      onkeydown={handleKeyDown}
      onfocus={() => isFocused = true}
      onblur={() => setTimeout(() => isFocused = false, 200)}
      {placeholder}
      class="w-full bg-transparent outline-none text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 dark:placeholder-zinc-500 font-normal"
      autocomplete="off"
      spellcheck="false"
    />

    <!-- Clear Button -->
    {#if value}
      <button
        type="button"
        onclick={clearSearch}
        aria-label="Clear query"
        class="p-1 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200 transition-colors rounded-full focus:outline-none"
      >
        <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    {/if}

    <!-- Search Submit Button -->
    <button
      type="button"
      onclick={() => value.trim() && onSearch(value.trim())}
      aria-label="Search"
      class="ml-2 px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded-full text-xs font-medium transition-colors shadow-xs"
    >
      Search
    </button>
  </div>

  <!-- Autocomplete Suggestions Dropdown -->
  {#if isFocused && suggestions.length > 0}
    <ul 
      class="absolute left-0 right-0 top-full mt-2 z-50 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-lg py-2 overflow-hidden animate-fade-in"
    >
      {#each suggestions as suggestion, index}
        <li>
          <button
            type="button"
            onmousedown={() => selectSuggestion(suggestion)}
            class="w-full text-left px-5 py-2.5 flex items-center gap-3 text-sm transition-colors {selectedIndex === index ? 'bg-zinc-100 dark:bg-zinc-800 text-blue-600 dark:text-blue-400 font-medium' : 'text-zinc-800 dark:text-zinc-200 hover:bg-zinc-50 dark:hover:bg-zinc-800/60'}"
          >
            <svg class="w-3.5 h-3.5 text-zinc-400 shrink-0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <span class="truncate">{suggestion}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>
