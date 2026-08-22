# Cadbri Search Engine — Full Bug & Issue Report

> **Audit Date:** 2026-05-03  
> **Stack:** React 19 + TypeScript + Vite + TailwindCSS 4 (Frontend) · Go/Chi (Backend)  
> **Method:** Full static code analysis · ESLint · TypeScript compiler · Line-by-line code review

---

## Summary Table

| # | File | Severity | Category | Description |
|---|------|----------|----------|-------------|
| 1 | `ResultsPage.tsx` | 🔴 High | Logic Bug | `effectiveLang` not in `useEffect`/`useCallback` deps — language changes don't re-search |
| 2 | `ResultsPage.tsx` | 🔴 High | Logic Bug | Infinite scroll `loadMore` still reads stale `searchParams` from closure |
| 3 | `ResultsPage.tsx` | 🔴 High | React Anti-pattern | 3× `setState` called synchronously inside `useEffect` — cascading re-renders |
| 4 | `ResultsPage.tsx` | 🟠 Medium | Logic Bug | Pagination "Next" always visible even when `hasMore = false` |
| 5 | `ResultsPage.tsx` | 🟠 Medium | Logic Bug | Result count stat line (`timeStr`) computed but never displayed |
| 6 | `ResultsPage.tsx` | 🟠 Medium | UX Bug | Duplicate `pb-8` on `<div>` wrapping results (double bottom padding) |
| 7 | `ResultList.tsx` | 🟠 Medium | Logic Bug | Non-image results hard-capped at 15 (`slice(0, 15)`) regardless of page |
| 8 | `ResultList.tsx` | 🟡 Low | Crash Risk | `result.parsed_url?.[1].split(...)` — optional chain only on index, not on `.split` |
| 9 | `ResultList.tsx` | 🟡 Low | Missing | `number_of_results` / `search_time` stat line never rendered |
| 10 | `ResultCard.tsx` | 🔴 High | Crash Risk | `new URL(result.url).hostname` throws if `result.url` is relative/malformed |
| 11 | `ResultCard.tsx` | 🟠 Medium | XSS Risk | `dangerouslySetInnerHTML={{ __html: result.content }}` — no sanitization |
| 12 | `SearchBar.tsx` | 🟡 Low | UX Bug | Voice search mic button does nothing (no Web Speech API hookup) |
| 13 | `SearchBar.tsx` | 🟡 Low | UX Bug | "ASK AI" button does nothing (placeholder with no handler) |
| 14 | `SearchBar.tsx` | 🟡 Low | Logic Bug | Arrow-up in autocomplete restores `initialQuery` but doesn't reset `selectedIndex` properly |
| 15 | `SearchTabs.tsx` | 🟠 Medium | Broken Feature | "Maps" tab sets `categories=map` but backend doesn't support `map` category |
| 16 | `SearchTabs.tsx` | 🟡 Low | Broken Feature | "Tools" button has no handler and does nothing |
| 17 | `Header.tsx` | 🟡 Low | UX Bug | Mobile hamburger menu button has no handler (no side menu implemented) |
| 18 | `Header.tsx` | 🟡 Low | UX Bug | "C" avatar button has no handler (not linked to profile/sign-in) |
| 19 | `SettingsModal.tsx` | 🟠 Medium | Broken Feature | "Homepage", "Search" tabs show "coming soon" — non-functional |
| 20 | `SettingsModal.tsx` | 🟠 Medium | Logic Bug | `selectedTheme` / `safeSearch` / `selectedRegionMode` are local state only — not persisted, not used |
| 21 | `SettingsModal.tsx` | 🟡 Low | UX Bug | `DropdownSelect` uses `window.innerWidth` SSR-style check that is fragile on resize |
| 15 | `SearchTabs.tsx` | 🟠 Medium | Broken Feature | "Maps" tab sets `categories=map` but backend doesn't support `map` category |
| 16 | `types.ts` | 🟡 Low | Missing Types | `ResultItem` missing `engine`, `engines`, `parsed_url`, `template` fields returned by backend |
| 17 | `ResultCard.tsx` | 🟡 Low | Enhancement | Engine tags (`result.engines`) are not rendered — users can't see source attribution |
| 18 | `ResultCard.tsx` | 🟡 Low | Enhancement | Published date (`result.publishedDate`) is not shown on results |
| 19 | `ImageResults.tsx` | 🟠 Medium | Missing Feature | Image resolution pill (e.g., "1920x1080") is not displayed on image cards |
| 20 | `ImageResults.tsx` | 🟠 Medium | Missing Feature | No source website domain shown on image cards |
| 21 | `ImageResults.tsx` | 🔴 High | UX Bug | Image click only opens thumbnail in new tab; no full-size preview modal/drawer |
| 22 | `VideoResults.tsx` | 🟠 Medium | Missing Feature | Video duration / length pill is not rendered on video thumbnails |
| 23 | `NewsResults.tsx` | 🟡 Low | Enhancement | News source publication and relative time ("2 hours ago") not formatted cleanly |
| 24 | `App.tsx` | 🔴 High | Critical Bug | Zero test coverage across entire frontend codebase (0 test files) |
| 25 | `index.html` | 🟡 Low | SEO/Accessibility | Missing `<meta name="description">` and `<meta name="theme-color">` tags |
| 26 | `index.html` | 🟡 Low | UX | Default favicon is Vite logo, not Cadbri brand icon |
| 27 | `App.tsx` | 🟡 Low | Bug | `window.scrollTo` called on initial render causing unnecessary scroll jump |
| 28 | `SearchTabs.tsx` | 🟡 Low | UX | Category icons are missing — only text labels are displayed |
| 29 | `api/client.ts` | 🟠 Medium | Bug | `fetchWithTimeout` abort controller leaks if response is stream-cancelled |
| 30 | `api/client.ts` | 🟠 Medium | Bug | Autocomplete URL uses `/api/autocompleter` but vite proxy rewrites `/api` → removes prefix, so it hits `/completer` on backend — path mismatch |
| 31 | `KnowledgePanel.tsx` | 🟡 Low | UX | Share button in KnowledgePanel does nothing |
| 32 | `Pagination.tsx` | 🔴 High | Logic Bug | Desktop "Next" button always rendered — no `hasMore` guard — clicks past end of results |
| 33 | Global | 🟠 Medium | Missing Feature | No `<meta>` description tag in results page (`document.title` set, but not meta desc) |
| 34 | Global | 🟡 Low | Missing | No `<title>` fallback / error boundary — if `ResultsPage` crashes, blank screen |
| 35 | Global | 🟡 Low | Accessibility | Multiple `<button>` elements without `type="button"` inside forms could accidentally submit |

---

## Detailed Bug Descriptions

### 🔴 BUG 1 — `effectiveLang` Missing from `useEffect` / `useCallback` Dependencies
**File:** `ResultsPage.tsx` — Lines 46, 86  
**Severity:** High — silent stale language  
**ESLint Rule:** `react-hooks/exhaustive-deps` (2× warnings reported)

```tsx
// Line 17
const effectiveLang = searchParams.get('language') || searchLang || undefined

// Line 29–46: effectiveLang NOT in deps array
useEffect(() => {
  doSearch({ q: query, categories: category, pageno: ..., language: effectiveLang })
}, [searchParams, doSearch, query, category])  // ← effectiveLang missing!

// Line 86: same in useCallback
}, [isFetchingMore, hasMore, query, category, page, searchParams])  // ← effectiveLang missing!
```

**Impact:** If the user changes their `searchLang` in Settings, the current result page will **never re-search with the new language** until `searchParams` changes. Searches on subsequent pages (`loadMore`) also use the stale language.

**Fix:**
```tsx
}, [searchParams, doSearch, query, category, effectiveLang])
// and in useCallback:
}, [isFetchingMore, hasMore, query, category, page, searchParams, effectiveLang])
```

---

### 🔴 BUG 2 — Stale `searchParams` in `loadMore` Closure
**File:** `ResultsPage.tsx` — Line 86  

`loadMore` has `searchParams` in its `useCallback` deps, which means every change to URL params creates a new function, which triggers the `IntersectionObserver` effect to re-subscribe. More critically, the language is read from `effectiveLang` (derived at render time) but missing from deps — so infinitely-loaded pages use stale language.

---

### 🔴 BUG 3 — `setState` Called Synchronously in `useEffect` (3 instances)
**File:** `ResultsPage.tsx` — Lines 30, 31, 32, 51, 52  
**ESLint Rule:** `react-hooks/set-state-in-effect` (3 errors)

```tsx
useEffect(() => {
  setAccumulatedResults([])  // ← sync setState in effect body
  setPage(1)                 // ← sync setState in effect body
  setHasMore(true)           // ← sync setState in effect body
  ...
}, [searchParams, doSearch, query, category])
```

This causes **3 extra renders** per search (one per `setState` call) before the actual search completes. React 18+ batches these in concurrent mode, but they're still anti-patterns per the rules and cause flickering.

---

### 🔴 BUG 4 — Pagination Desktop "Next" Button Has No `hasMore` Guard
**File:** `Pagination.tsx` — Lines 81–87  

```tsx
<button 
  onClick={() => goToPage(currentPage + 1)}
  className="..."
>
  Next  // ← ALWAYS visible, no disabled state, no hasMore check
</button>
```

The "Previous" button correctly checks `currentPage > 1` before rendering, but "Next" **always renders** even when there are no more results. Users can click "Next" forever going to empty result pages.

**Fix:**
```tsx
{hasMore && (
  <button onClick={() => goToPage(currentPage + 1)} ...>Next</button>
)}
```

---

### 🔴 BUG 5 — `new URL(result.url).hostname` Throws on Malformed URLs
**File:** `ResultCard.tsx` — Line 17  

```tsx
const domain = result.parsed_url?.[1] || new URL(result.url).hostname
```

If `parsed_url` is missing AND `result.url` is a relative URL or otherwise malformed, `new URL()` throws a `TypeError`, **crashing the entire ResultCard** (and potentially the whole results list without an error boundary).

**Fix:**
```tsx
let domain = result.parsed_url?.[1] || ''
if (!domain) {
  try { domain = new URL(result.url).hostname } catch { domain = result.url }
}
```

---

### 🔴 BUG 6 — XSS Risk via `dangerouslySetInnerHTML` Without Sanitization
**File:** `ResultCard.tsx` — Lines 67, 124, 183, 203; `ResultList.tsx` — Line 62  

```tsx
<span dangerouslySetInnerHTML={{ __html: result.content }} />
```

The `content` field comes directly from upstream search engines. If the backend returns unexpected HTML (e.g., `<script>`, event handlers), this is a **stored XSS vector**.

**Fix:** Use a sanitizer or text node rendering:
```tsx
import DOMPurify from 'dompurify'
<span dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(result.content) }} />
```

---

### 🟠 BUG 7 — Non-Image Results Hard-Capped at 15 Per Page
**File:** `ResultList.tsx` — Line 107  

```tsx
{data.results.slice(0, 15).map((result, i) => (
```

When `loadMore` appends additional results to `accumulatedResults`, those results flow through `displayData` and into `ResultList`. But `.slice(0, 15)` silently discards everything beyond 15 results, so **infinite scroll / pagination is completely broken for web/news/video results** — only images use the full array.

**Fix:** Remove the slice, or make it configurable:
```tsx
{data.results.map((result, i) => (
```

---

### 🟠 BUG 8 — Autocomplete Proxy Path Mismatch
**File:** `api/cadbri.ts` — Line 192  
**File:** `vite.config.ts` — Lines 17–21  

```ts
// In cadbri.ts:
const url = new URL('/api/autocompleter', window.location.origin)

// In vite.config.ts:
'/api': {
  target: 'http://localhost:2222',
  rewrite: (path) => path.replace(/^\/api/, ''),  // removes /api prefix
}
```

---

### 🟠 BUG 9 — `search_time` / Result Count Never Displayed
**Files:** `ResultList.tsx` — Lines 52–53; `ResultsPage.tsx`  

```tsx
// ResultList.tsx
const formattedCount = new Intl.NumberFormat('en-US').format(data.number_of_results)
const timeStr = data.search_time ? ` (${data.search_time.toFixed(2)} seconds)` : ''
```

Both `formattedCount` and `timeStr` are computed but **never rendered** — the stat line "About 1,540,000 results (0.84 seconds)" is completely missing from the UI. ESLint flags `timeStr` as unused.

---

### 🟠 BUG 10 — Maps Tab Sends Unsupported Category
**File:** `SearchTabs.tsx` — Line 24  

```tsx
{ id: 'map', label: t('tabs.maps'), icon: TAB_ICONS.map },
```

When clicked, this sets `categories=map` in the URL. Ensure backend supports `map` category for OpenStreetMap.

---

### 🟠 BUG 11 — Settings State Not Persisted
**File:** `SettingsModal.tsx` — Lines 70–71, 198  

```tsx
const [selectedTheme, setSelectedTheme] = useState('System')
const [selectedRegionMode, setSelectedRegionMode] = useState('regionalDisabled')
const [safeSearch, setSafeSearch] = useState('Strict')
```

These 3 settings are **local component state only** — they are not saved to `localStorage`, not passed to `SettingsContext`, and not used anywhere in the app. Every time the settings modal opens, they reset to defaults. Theme/dark mode is completely non-functional.

---

### 🟠 BUG 12 — `SettingsContext` Exported Function Breaks Fast Refresh
**File:** `SettingsContext.tsx` — Line 251  
**ESLint Rule:** `react-refresh/only-export-components`  

```tsx
// Both of these are exported from the same file:
export function SettingsProvider(...) { ... }  // component
export function useSettings() { ... }           // non-component hook
```

This breaks Vite's React Fast Refresh — changes to this file will cause a **full page reload** instead of hot module replacement, making development slower.

---

### 🟡 BUG 13 — Autocomplete Key Collision
**File:** `Autocomplete.tsx` — Line 40  

```tsx
<li key={s} ...>
```

If the autocomplete API returns duplicate suggestions (e.g., `["react", "react"]`), React will throw a duplicate key warning and may render incorrectly.

**Fix:** Use index: `key={`${s}-${i}`}`

---

### 🟡 BUG 14 — Optional Chain Not Deep Enough (Potential Crash)
**File:** `ResultList.tsx` — Line 99  

```tsx
<span>{result.parsed_url?.[1].split('.')[0] || 'Site'}</span>
```

The optional chain `?.` only protects the array index access. If `parsed_url[1]` is defined but is somehow not a string (backend inconsistency), `.split` will throw.

**Fix:**
```tsx
result.parsed_url?.[1]?.split('.')[0] || 'Site'
```

---

### 🟡 BUG 15 — Mobile Hamburger Menu Has No Handler
**File:** `Header.tsx` — Line 44  

```tsx
<button className="p-2 -ml-2 text-gray-600 cursor-pointer" aria-label="Menu">
  ...
</button>
```

No `onClick` handler — the mobile menu button is **completely non-functional**. No drawer/sidebar component exists.

---

### 🟡 BUG 16 — Voice Search & ASK AI Buttons Are Non-Functional
**File:** `SearchBar.tsx` — Lines 133, 142  

Both buttons are rendered with no `onClick` handlers. The Voice Search button has no Web Speech API integration. ASK AI is a completely blank placeholder.

---

### 🟡 BUG 17 — Footer Imports `Link` but Never Uses It
**File:** `Footer.tsx` — Line 2  

```tsx
import { Link } from 'react-router-dom'
// ...
<footer className="w-full"></footer>  // empty — Link never used
```

Unused import warning. The footer is also **completely empty** — no privacy policy, about, or attribution links.

---

### 🟡 BUG 18 — `DropdownSelect` Uses `window.innerWidth` Synchronously
**File:** `SettingsModal.tsx` — Line 21  

```tsx
style={{ minWidth: typeof window !== 'undefined' && window.innerWidth < 768 ? '100%' : minWidth }}
```

This runs once at **render time**, not on resize. If the user resizes the window, the style won't update. Should use a CSS `@media` rule or a `useWindowSize` hook.

---

## Summary by File

| File | Errors | Warnings |
|------|--------|----------|
| `ResultsPage.tsx` | 3 setState-in-effect, 1 missing dep | 2 missing deps |
| `SettingsContext.tsx` | 4 empty catch, 1 setState-in-effect, 1 fast-refresh | — |
| `ResultCard.tsx` | 1 crash risk (URL parse), 1 XSS | — |
| `ResultList.tsx` | 1 optional chain crash, 1 results cap | 1 unused var |
| `Pagination.tsx` | 1 Next button always shown | — |
| `Autocomplete.tsx` | 1 key collision | — |
| `SearchBar.tsx` | 2 dead buttons | — |
| `Header.tsx` | 1 dead button, 1 dead avatar | — |
| `SettingsModal.tsx` | 3 unpersisted states, 1 z-index conflict | — |
| `Footer.tsx` | 1 unused import | — |
| `api/client.ts` | 1 proxy path conflict | — |
| `SearchTabs.tsx` | 1 invalid category, 1 dead Tools button | — |

---

## Recommended Fix Priority

### Fix Immediately (Bugs that break core functionality):
1. **Bug 7** — Remove `slice(0, 15)` — pagination/infinite scroll doesn't work
2. **Bug 4** — Add `hasMore` guard to Pagination "Next" button
3. **Bug 1** — Add `effectiveLang` to `useEffect` deps
4. **Bug 5** — Display the result count / search time stat line
5. **Bug 10** — Fix or remove the Maps tab (use a valid category or redirect to an external map)

### Fix Soon (Quality / Stability):
6. **Bug 5** — Wrap `new URL()` in try/catch in `ResultCard`
7. **Bug 11** — Persist theme, safeSearch, regionMode to localStorage/context
8. **Bug 6** — Add DOMPurify for `dangerouslySetInnerHTML`
9. **Bug 12** — Split `SettingsContext.tsx` to fix fast refresh

### Nice to Have:
10. **Bug 15/16/17** — Implement mobile menu, voice search, or remove buttons
11. **Bug 9** — Fix `timeStr` variable — either use it or remove it
12. **Bug 13** — Fix autocomplete key collision
13. **Bug 18** — Fix responsive dropdown sizing with CSS
