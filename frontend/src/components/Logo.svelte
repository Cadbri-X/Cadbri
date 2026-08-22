<script lang="ts">
  export let size: 'sm' | 'md' | 'lg' | 'xl' = 'md';
  export let showText = false;
  export let textClass = '';

  const iconDimensions = {
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-16 h-16',
    xl: 'w-20 h-20'
  };

  const textSizes = {
    sm: 'text-lg',
    md: 'text-xl',
    lg: 'text-3xl sm:text-4xl',
    xl: 'text-4xl sm:text-5xl'
  };
</script>

<div class="flex items-center gap-2.5 select-none">
  <!-- Dynamic Vector SVG Logomark -->
  <div class="{iconDimensions[size]} shrink-0 transition-transform duration-200 group-hover:scale-105">
    <svg 
      viewBox="0 0 48 48" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      class="w-full h-full drop-shadow-sm"
    >
      <defs>
        <!-- Primary Brand Gradient -->
        <linearGradient id="cadbri-gradient-1" x1="4" y1="4" x2="44" y2="44" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stop-color="#38bdf8" />
          <stop offset="45%" stop-color="#2563eb" />
          <stop offset="100%" stop-color="#1d4ed8" />
        </linearGradient>

        <!-- Accent Glow Gradient -->
        <linearGradient id="cadbri-gradient-2" x1="16" y1="8" x2="44" y2="36" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stop-color="#60a5fa" />
          <stop offset="100%" stop-color="#0284c7" />
        </linearGradient>

        <linearGradient id="cadbri-lens-grad" x1="16" y1="16" x2="32" y2="32" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stop-color="#38bdf8" />
          <stop offset="100%" stop-color="#2563eb" />
        </linearGradient>

        <!-- Drop Shadow Filter -->
        <filter id="cadbri-glow" x="-10%" y="-10%" width="120%" height="120%" filterUnits="userSpaceOnUse">
          <feDropShadow dx="0" dy="2" stdDeviation="2.5" flood-color="#2563eb" flood-opacity="0.3" />
        </filter>
      </defs>

      <!-- Outer Squircle / Base Capsule with Subtle Depth -->
      <rect 
        x="2" 
        y="2" 
        width="44" 
        height="44" 
        rx="13" 
        fill="url(#cadbri-gradient-1)" 
        filter="url(#cadbri-glow)"
      />

      <!-- Dynamic Futuristic 'C' & Lens Aperture -->
      <!-- Main 'C' Arc Path -->
      <path 
        d="M 33 13.5 C 30.5 11.5 27.2 10 23.5 10 C 15.5 10 9 16.5 9 24.5 C 9 32.5 15.5 39 23.5 39 C 27.5 39 31 37.3 33.5 35" 
        stroke="white" 
        stroke-width="3.6" 
        stroke-linecap="round" 
      />

      <!-- Speed Rays Emanating from the 'C' -->
      <path 
        d="M 30 18 L 38 12" 
        stroke="white" 
        stroke-width="3" 
        stroke-linecap="round" 
        opacity="0.9"
      />
      <path 
        d="M 34 23 L 41 19" 
        stroke="white" 
        stroke-width="2.5" 
        stroke-linecap="round" 
        opacity="0.75"
      />

      <!-- Core Search Eye / Focus Aperture -->
      <circle 
        cx="23.5" 
        cy="24.5" 
        r="6.5" 
        fill="white" 
      />
      <circle 
        cx="23.5" 
        cy="24.5" 
        r="4" 
        fill="url(#cadbri-gradient-1)" 
      />
      <circle 
        cx="22" 
        cy="23" 
        r="1.2" 
        fill="white" 
        opacity="0.9"
      />
    </svg>
  </div>

  <!-- Optional Text -->
  {#if showText}
    <span class="font-bold tracking-tight text-zinc-900 dark:text-white {textSizes[size]} {textClass}">
      Cadbri
    </span>
  {/if}
</div>
