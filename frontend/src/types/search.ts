export interface SearchResult {
  url: string;
  title: string;
  content: string;
  parsed_url?: string[];
  engines?: string[];
  category?: string;
  thumbnail?: string;
  img_src?: string;
  publishedDate?: string;
  pubdate?: string;
  author?: string;
  views?: string;
  length?: number;
  iframe_src?: string;
  audio_src?: string;
  metadata?: string;
  score?: number;
}

export interface Answer {
  url?: string;
  answer: string;
  engine?: string;
  template?: string;
}

export interface InfoboxURL {
  title: string;
  url: string;
}

export interface InfoboxAttribute {
  label: string;
  value: string;
}

export interface Infobox {
  infobox: string;
  id?: string;
  content?: string;
  img_src?: string;
  engine?: string;
  urls?: InfoboxURL[];
  attributes?: InfoboxAttribute[];
}

export interface SearchResponse {
  query: string;
  number_of_results: number;
  results: SearchResult[];
  answers?: Answer[];
  corrections?: string[];
  infoboxes?: Infobox[];
  suggestions?: string[];
  unresponsive_engines?: string[][];
}

export type CategoryId = 'general' | 'images' | 'videos' | 'news' | 'map' | 'it' | 'science';

export interface CategoryTab {
  id: CategoryId;
  label: string;
  icon: string;
  engines?: string[];
}

export interface SearchFilters {
  timeRange: '' | 'day' | 'week' | 'month' | 'year';
  safeSearch: number; // 0: None, 1: Moderate, 2: Strict
  language: string;
}
