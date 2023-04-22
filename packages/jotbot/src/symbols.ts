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
] as const

export const globalSymbols = ['func', 'var', 'class', 'iface'] as const

export const ownerSymbols = ['class', 'iface'] as const

export type SymbolType = (typeof symbolTypes)[number]

export type GlobalSymbol = (typeof globalSymbols)[number]

export type OwnerSymbol = (typeof ownerSymbols)[number]

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
