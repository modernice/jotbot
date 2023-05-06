import { describe, it } from 'vitest'
import { createFinder } from '../src/finder'
import { SymbolType, symbolTypes } from '../src/symbols'
import { expectFindings } from './testutils'
import { heredoc } from '../src/utils'
import { RawIdentifier } from '../src'

describe('finder', () => {
  it('finds uncommented types', () => {
    const code = heredoc`
      export var foo = 'foo'
      export const bar = 'bar'
      export function baz() {}
      export const foobar = () => {}
      export interface Foo {
        foo: string
        bar(): string
        baz: () => string
      }
      export class Bar {
        foo = 'foo'
        bar() { return 'bar' }
        baz = () => 'baz'
      }
      export type Baz = {
        foo: string
        bar(): number
        baz: () => boolean
      }
    `
    const { find } = createFinder()

    const findings = find(code)

    expectFindings(findings, [
      'var:foo',
      'var:bar',
      'func:baz',
      'var:foobar',
      'iface:Foo',
      'prop:Foo.foo',
      'method:Foo.bar',
      'prop:Foo.baz',
      'class:Bar',
      'prop:Bar.foo',
      'method:Bar.bar',
      'prop:Bar.baz',
      'type:Baz',
      'prop:Baz.foo',
      'method:Baz.bar',
      'prop:Baz.baz',
    ])
  })

  it("doesn't find unexported symbols", () => {
    const code = heredoc`
      var foo = 'foo'
      const bar = 'bar'
      function baz() {}
      const foobar = () => {}
      interface Foo {
        foo: string
      }
      class Bar {
        foo = 'foo'
      }
      export const hello = 'hello'
    `

    const { find } = createFinder()

    const findings = find(code)

    expectFindings(findings, ['var:hello'])
  })
})

describe(`'symbols' option`, () => {
  const code = heredoc`
    export var foo = 'foo'
    export const bar = 'bar'
    export function baz() {}
    export const foobar = () => {}
    export interface Foo {
      foo: string
      bar(): string
      baz: () => string
    }
    export class Bar {
      foo = 'foo'
      bar() { return 'bar' }
      baz = () => 'baz'
    }
  `

  const all = [
    'var:foo',
    'var:bar',
    'func:baz',
    'var:foobar',
    'iface:Foo',
    'prop:Foo.foo',
    'method:Foo.bar',
    'prop:Foo.baz',
    'class:Bar',
    'prop:Bar.foo',
    'method:Bar.bar',
    'prop:Bar.baz',
  ] as RawIdentifier[]

  function filter(symbols: SymbolType[]) {
    return all.filter((id) => symbols.some((sym) => id.startsWith(sym)))
  }

  const tests = [
    {
      name: 'finds all symbols by default',
      symbols: [],
      expected: all,
    },
    {
      name: 'finds all symbols when all are specified',
      symbols: symbolTypes,
      expected: all,
    },
    {
      name: 'finds only specified symbols (1)',
      symbols: ['var', 'iface', 'class'],
      expected: filter(['var', 'iface', 'class']),
    },
    {
      name: 'finds only specified symbols (2)',
      symbols: ['prop', 'method'],
      expected: filter(['prop', 'method']),
    },
  ] as const

  for (const tt of tests) {
    it(tt.name, () => {
      const { find } = createFinder({ symbols: tt.symbols })
      const findings = find(code)

      expectFindings(findings, tt.expected)
    })
  }
})
