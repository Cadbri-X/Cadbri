<script lang="ts">
  import Logo from './Logo.svelte';
  import SearchBar from './SearchBar.svelte';
  import type { CategoryId } from '../types/search';

  let { onSearch, isDarkMode = $bindable(false) } = $props();

  let searchQuery = $state('');

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

<div class="min-h-screen flex flex-col justify-between bg-white dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 transition-colors">
  <!-- Top Navigation Header -->
  <header class="flex items-center justify-between px-6 py-4">
    <div class="flex items-center gap-3">
      <span class="text-xs font-semibold uppercase tracking-wider text-zinc-400 dark:text-zinc-500">
        Privacy-First Search
      </span>
    </div>

    <div class="flex items-center gap-2">
      <!-- Dark Mode Button -->
      <button
        type="button"
        onclick={toggleDarkMode}
        aria-label="Toggle theme"
        class="p-2 rounded-full text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors cursor-pointer"
      >
        {#if isDarkMode}
          <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg>
        {:else}
          <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>
        {/if}
      </button>
    </div>
  </header>

  <!-- Centered Search Content -->
  <main class="flex-1 flex flex-col items-center justify-center px-4 max-w-3xl mx-auto w-full -mt-12">
    <!-- Minimalist Logo & Brand Title Side-by-Side -->
    <div class="flex flex-col items-center mb-8 animate-fade-in text-center">
      <div class="flex items-center gap-3.5 sm:gap-4 mb-2">
        <Logo size="lg" />
        <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-zinc-900 dark:text-white">
          Cadbri
        </h1>
      </div>
      <p class="text-xs sm:text-sm text-zinc-500 dark:text-zinc-400 font-medium">
        Ultra-Fast, Minimalist Metasearch Engine
      </p>
    </div>

    <!-- Centered Search Bar -->
    <div class="w-full max-w-xl">
      <SearchBar 
        bind:value={searchQuery} 
        onSearch={(q) => onSearch(q, 'general')}
        size="large"
        autoFocus={true}
      />
    </div>

    <!-- Privacy Badges -->
    <div class="flex items-center gap-6 mt-10 text-xs text-zinc-400 dark:text-zinc-500">
      <span class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5 text-blue-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        No Tracking
      </span>
      <span class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5 text-emerald-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
        65 Engines Aggregated
      </span>
      <span class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5 text-purple-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m10 15 5-3-5-3v6Z"/></svg>
        Sub-Second Speed
      </span>
    </div>
  </main>

  <!-- Minimalist Footer -->
  <footer class="w-full border-t border-zinc-100 dark:border-zinc-900 px-6 py-4 flex flex-col sm:flex-row items-center justify-between text-xs text-zinc-400 dark:text-zinc-500 gap-2">
    <div class="flex items-center gap-4">
      <span>Powered by Cadbri</span>
    </div>
    <div class="flex items-center gap-4">
      <a href="https://github.com/Cadbri-X/Cadbri" target="_blank" rel="noopener noreferrer" class="hover:underline text-blue-600 dark:text-blue-400">
        GitHub
      </a>
      <span>•</span>
      <span>Lightweight & Minimalist</span>
    </div>
  </footer>
</div>
