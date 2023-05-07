/**
 * The symbol types to look for in the source code.
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
 * An array of symbol types to look for in the source code, including functions,
 * variables, classes, interfaces, and types. The "globalSymbols" variable is a
 * subset of these symbol types that are considered to be global symbols and
 * includes functions, variables, classes, interfaces, and types.
 */
export const globalSymbols = ['func', 'var', 'class', 'iface', 'type'] as const

/**
 * Represents the possible types of symbols that can be found in source code.
 * Includes constants for function, variable, class, method, interface,
 * property, and type symbols. The type can be checked using the `isSymbol`
 * function and configured using the `configureSymbols` function with a readonly
 * array of desired symbol types.
 */
export type SymbolType = (typeof symbolTypes)[number]

/**
 * GlobalSymbol is a type alias that represents a specific set of symbol types
 * to look for in source code. It is defined as a subset of the symbolTypes
 * constant, and includes only 'func', 'var', 'class', 'iface', and 'type'. The
 * purpose of GlobalSymbol is to provide a focused set of symbols for use in
 * configuring symbol search behavior via the configureSymbols function.
 */
export type GlobalSymbol = (typeof globalSymbols)[number]

/**
 * Checks if the provided input is a valid symbol type by comparing it against a
 * list of recognized {@link SymbolType}s. Returns a boolean indicating whether
 * the input is a valid symbol type or not.
 */
export function isSymbol(s: unknown): s is SymbolType {
  return symbolTypes.includes(s as any)
}

/**
 * Configures the symbol types to look for in the source code. The function
 * takes an array of symbol types and returns either the provided array or, if
 * empty, a default array of symbol types.
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
