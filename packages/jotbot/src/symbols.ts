/**
 * Represents the collection of specific symbol type identifiers used within the
 * system, encompassing items such as functions, variables, classes, methods,
 * interfaces, properties, and custom types. Each member of this array is a
 * literal type representing a distinct kind of symbol that can be analyzed or
 * manipulated by the system's features.
 */
export const symbolTypes = [
  'func',
  'var',
  'class',
  'method',
  'iface',
  'prop',
  'type',
] as const

/**
 * Maintains a collection of symbol types that are considered global across the
 * environment. These symbol types include fundamental constructs such as
 * functions, variables, classes, interfaces, and types. Each member of
 * `globalSymbols` is a string literal corresponding to these constructs and can
 * be used in conjunction with type operations to reference global symbol types
 * as defined by {@link GlobalSymbol}.
 */
export const globalSymbols = ['func', 'var', 'class', 'iface', 'type'] as const

export type SymbolType = (typeof symbolTypes)[number]

export type GlobalSymbol = (typeof globalSymbols)[number]

/**
 * Determines whether a given value is a recognized symbol type. This check is
 * performed against an array of predefined symbol types. If the value matches
 * one of the symbol types, the function returns `true`, indicating that the
 * value is a valid {@link SymbolType}. Otherwise, it returns `false`. The
 * function employs a type predicate to inform TypeScript's type checker of the
 * result.
 */
export function isSymbol(s: unknown): s is SymbolType {
  return symbolTypes.includes(s as any)
}

/**
 * Configures a provided array of symbol types, returning either the default set
 * of symbol types if the input is empty or the provided array otherwise. It
 * ensures that the returned value is of a type consistent with either the full
 * list of symbol types or a subset specified by the caller. If an empty array
 * is provided, `configureSymbols` returns the full list of default symbol
 * types; otherwise, it returns the input array as is. This function is generic
 * and can be tailored to accept a specific subset of symbol types, enforcing
 * compile-time checks on the input symbols.
 */
export function configureSymbols<const Symbols extends readonly SymbolType[]>(
  symbols: Symbols,
): Symbols['length'] extends 0 ? typeof symbolTypes : Symbols {
  return (
    isEmpty(symbols) ? symbolTypes : symbols
  ) as Symbols['length'] extends 0 ? typeof symbolTypes : Symbols
}

function isEmpty<T>(
  array: readonly T[],
): array is readonly T[] & { length: 0 } {
  return array.length === 0
}
