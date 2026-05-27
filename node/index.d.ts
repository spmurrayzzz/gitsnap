export interface Options {
  worktree?: string;
}

export interface SaveOptions extends Options {
  alias?: string;
}

export interface RestoreOptions extends Options {
  paths?: string[];
}

export interface Alias {
  name: string;
  hash: string;
  updated_at: string;
}

export function init(options?: Options): void;
export function cleanup(options?: Options): void;
export function save(options?: SaveOptions): string;
export function resolve(ref: string, options?: Options): string;
export function diff(ref: string, options?: Options): string;
export function files(ref: string, options?: Options): string[];
export function restore(ref: string, options?: RestoreOptions): void;
export function aliases(options?: Options): Alias[];
