import type { SearchResponse, CategoryId, SearchFilters } from '../types/search';

const API_BASE = '/api';

export async function searchCadbri(
  query: string,
  category: CategoryId = 'general',
  page: number = 1,
  filters?: Partial<SearchFilters>,
  signal?: AbortSignal
): Promise<SearchResponse> {
  const params = new URLSearchParams({
    q: query,
    format: 'json',
    pageno: page.toString(),
  });

  if (category !== 'general') {
    params.set('categories', category);
  }

  if (filters?.timeRange) {
    params.set('time_range', filters.timeRange);
  }

  if (filters?.safeSearch !== undefined) {
    params.set('safesearch', filters.safeSearch.toString());
  }

  if (filters?.language && filters.language !== 'all') {
    params.set('language', filters.language);
  }

  const url = `${API_BASE}/search?${params.toString()}`;
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Accept': 'application/json',
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`Search request failed with status: ${response.status}`);
  }

  return response.json();
}

export async function fetchSuggestions(
  query: string,
  signal?: AbortSignal
): Promise<string[]> {
  const trimmed = query.trim();
  if (!trimmed) return [];

  const url = `/autocompleter?q=${encodeURIComponent(trimmed)}`;
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
      signal,
    });

    if (!response.ok) return [];

    const data = await response.json();
    if (Array.isArray(data)) {
      if (Array.isArray(data[1])) {
        return data[1];
      }
      return data.filter((item): item is string => typeof item === 'string');
    }
    return [];
  } catch {
    return [];
  }
}
