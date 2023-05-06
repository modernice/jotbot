import type ts from 'typescript'
import { expect } from 'vitest'
import { RawIdentifier } from '../src/identifier'
import { SymbolType } from '../src'

export function tryOut<Fn extends (...args: any[]) => any>(
  fn: Fn,
): [ReturnType<Fn>, null] | [null, Error] {
  try {
    return [fn(), null]
  } catch (err) {
    const error = err instanceof Error ? err : new Error(String(err))
    return [null, error]
  }
}

export function expectFindings<Symbols extends SymbolType = SymbolType>(
  got: RawIdentifier<Symbols>[],
  want: (RawIdentifier<Symbols> | RawIdentifier<Symbols>)[],
) {
  const _want = want.map(
    (item): RawIdentifier<Symbols> =>
      isRawIdentifier<Symbols>(item) ? (item as RawIdentifier<Symbols>) : item,
  )

  _want.sort()
  got.sort()

  expect(got).toEqual(_want)
}

export function debugNode(n: ts.Node, msg?: string) {
  console.log(`\n${msg ?? 'Node:'}\n${n.getText()}`)
}

function isRawIdentifier<Symbols extends SymbolType = SymbolType>(
  identifier: unknown,
): identifier is RawIdentifier<Symbols> {
  return typeof identifier === 'string'
}
