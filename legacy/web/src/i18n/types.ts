// The shape of a dictionary, derived from fr.ts — the source of truth.
import type { fr } from "./fr";

export type Locale = "fr" | "en";

/** A message that depends on a count: Intl.PluralRules picks the form. */
export type Plural = { readonly one: string; readonly other: string };

type Dict = typeof fr;

/** Every message of the interface. */
export type Key = keyof Dict;

/** A dictionary: the same keys as fr.ts, each a string or a plural. */
export type Messages = {
  readonly [K in Key]: Dict[K] extends string ? string : Plural;
};

type Holes<S> = S extends `${string}{${infer P}}${infer R}`
  ? P | Holes<R>
  : never;

/** The {placeholders} of a message; a plural always takes "count". */
export type Placeholder<K extends Key> = Dict[K] extends string
  ? Holes<Dict[K]>
  : Dict[K] extends Plural
    ? Holes<Dict[K]["one"]> | Holes<Dict[K]["other"]> | "count"
    : never;

export type Params<K extends Key> = {
  [P in Placeholder<K>]: P extends "count" ? number : string | number;
};

/** A message without placeholders takes no argument; one with placeholders requires them all. */
export type Args<K extends Key> = [Placeholder<K>] extends [never]
  ? [params?: undefined]
  : [params: Params<K>];
