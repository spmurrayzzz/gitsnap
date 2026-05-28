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

export function init(options?: Options): Promise<void>;
export function cleanup(options?: Options): Promise<void>;
export function save(options?: SaveOptions): Promise<string>;
export function resolve(ref: string, options?: Options): Promise<string>;
export function diff(ref: string, options?: Options): Promise<string>;
export function files(ref: string, options?: Options): Promise<string[]>;
export function restore(ref: string, options?: RestoreOptions): Promise<void>;
export function aliases(options?: Options): Promise<Alias[]>;

export namespace sync {
  function init(options?: Options): void;
  function cleanup(options?: Options): void;
  function save(options?: SaveOptions): string;
  function resolve(ref: string, options?: Options): string;
  function diff(ref: string, options?: Options): string;
  function files(ref: string, options?: Options): string[];
  function restore(ref: string, options?: RestoreOptions): void;
  function aliases(options?: Options): Alias[];
}
