import type ts from 'typescript'
import { expect } from 'vitest'
import {
  NaturalLanguageTarget,
  RawIdentifier,
  describeIdentifier,
} from '../src/identifier'
import type { Finding } from '../src/finder'
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
  got: Finding<Symbols>[],
  want: (RawIdentifier<Symbols> | Finding<Symbols>)[],
) {
  const _want = want.map(
    (item): Finding<Symbols> =>
      isRawIdentifier<Symbols>(item)
        ? {
            identifier: item as RawIdentifier<Symbols>,
            target: describeIdentifier(item) as NaturalLanguageTarget<Symbols>,
          }
        : item,
  )

  _want.sort((a, b) => (a.target <= b.target ? -1 : 1))
  got.sort((a, b) => (a.target <= b.target ? -1 : 1))

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
