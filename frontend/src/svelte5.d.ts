// Svelte 5 / Svelte JSX / Vite Ambient Typings for IDE Language Server

declare namespace svelteHTML {
  interface HTMLAttributes<T extends EventTarget = any> {
    [key: string]: any;
  }
}

declare module 'vite' {
  export function defineConfig(config: any): any;
}

declare module '@sveltejs/vite-plugin-svelte' {
  export function svelte(options?: any): any;
  export function vitePreprocess(options?: any): any;
}

declare module '@tailwindcss/vite' {
  export default function tailwindcss(options?: any): any;
}

declare module 'svelte' {
  export class SvelteComponent<
    Props extends Record<string, any> = Record<string, any>,
    Events extends Record<string, any> = any,
    Slots extends Record<string, any> = any
  > {
    $$prop_def: Props;
    $$events_def: Events;
    $$slots_def: Slots;
    $$slot_def: Slots;
    $set(props?: Partial<Props>): void;
    $on(event: string, callback: (event: any) => void): () => void;
    $destroy(): void;
    constructor(options?: any);
    [key: string]: any;
  }

  export class SvelteComponentTyped<
    Props extends Record<string, any> = Record<string, any>,
    Events extends Record<string, any> = any,
    Slots extends Record<string, any> = any
  > extends SvelteComponent<Props, Events, Slots> {}

  export interface Component<
    Props extends Record<string, any> = Record<string, any>,
    Exports extends Record<string, any> = Record<string, any>,
    Bindings extends string = string
  > {
    (props: Props): Exports;
    element?: HTMLElement;
  }

  export type ComponentType<Comp extends SvelteComponent = SvelteComponent> = new (...args: any[]) => Comp;

  export const mount: (component: any, options: { target: Element | Document | ShadowRoot; props?: any }) => any;
  export const unmount: (component: any) => void;
  export const onMount: (fn: () => void | (() => void)) => void;
  export const onDestroy: (fn: () => void) => void;
  export const tick: () => Promise<void>;
}

declare module '*.svelte' {
  import { SvelteComponent } from 'svelte';
  export default class SvelteComponentTyped<
    Props extends Record<string, any> = Record<string, any>,
    Events extends Record<string, any> = any,
    Slots extends Record<string, any> = any
  > extends SvelteComponent<Props, Events, Slots> {
    constructor(options?: any);
  }
}

// Global Svelte 5 Runes
declare function $state<T = any>(initial?: T): T;
declare namespace $state {
  function raw<T = any>(initial?: T): T;
  function snapshot<T = any>(state: T): T;
}

declare function $derived<T = any>(expression: T): T;
declare namespace $derived {
  function by<T = any>(fn: () => T): T;
}

declare function $props<T = any>(): T;
declare function $bindable<T = any>(fallback?: T): T;
declare function $effect(fn: () => void | (() => void)): void;
declare namespace $effect {
  function pre(fn: () => void | (() => void)): void;
  function root(fn: () => void | (() => void)): () => void;
}
declare function $inspect(...values: any[]): { with: (fn: (...args: any[]) => void) => void };
declare function $host(): any;
