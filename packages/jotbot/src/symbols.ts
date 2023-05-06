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

export const globalSymbols = ['func', 'var', 'class', 'iface', 'type'] as const

export type SymbolType = (typeof symbolTypes)[number]

export type GlobalSymbol = (typeof globalSymbols)[number]

export function isSymbol(s: unknown): s is SymbolType {
  return symbolTypes.includes(s as any)
}

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
