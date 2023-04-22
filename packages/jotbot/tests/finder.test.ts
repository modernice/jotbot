import { describe, expect, it } from 'vitest'
import { isEqual } from 'lodash'
import {
  createFinder,
  printFindings,
  removeNodesFromFindings,
} from '../src/finder'
import { symbolTypes } from '../src/symbols'
import { fixtureRoot } from './fixtures'
import {
  expectFiles,
  expectFindings,
  expectNoFiles,
  expectNotFound,
  without,
} from './testutils'

describe('finder', () => {
  it('finds uncommented types', () => {
    const root = fixtureRoot('basic')
    const { findUncommented } = createFinder(root)

    const nodes = findUncommented()

    expectFindings(nodes, {
      'foo.ts': ['var:foobar', 'func:foo'],
      'bar.ts': ['class:Bar', 'method:Bar.bar'],
      'baz/baz.ts': ['var:baz', 'func:foobar'],
    })
  })

  it("doesn't find unexported types", () => {
    const root = fixtureRoot('unexported')
    const { findUncommented } = createFinder(root)

    const nodes = findUncommented()

    expectFindings(nodes, {
      'foo.ts': ['class:Foo', 'func:baz'],
    })

    expectNotFound(nodes, ['class:Bar', 'func:baz.foobar', 'method:Foo.foo'])
  })

  it('supports glob include patterns', () => {
    const root = fixtureRoot('tree')
    const { findUncommented } = createFinder(root, {
      include: ['**/foo.ts'],
    })

    const nodes = findUncommented()

    expectFiles(nodes, [
      'foo/foo/foo.ts',
      'foo/bar/foo.ts',
      'foo/baz/foo.ts',
      'bar/foo/foo.ts',
      'bar/bar/foo.ts',
      'bar/baz/foo.ts',
      'baz/foo/foo.ts',
      'baz/bar/foo.ts',
      'baz/baz/foo.ts',
    ])
  })

  it('supports multiple glob include patterns', () => {
    const root = fixtureRoot('tree')
    const { findUncommented } = createFinder(root, {
      include: ['**/foo.ts', '**/bar.ts'],
    })

    const nodes = findUncommented()

    expectFiles(nodes, [
      'foo/foo/foo.ts',
      'foo/bar/foo.ts',
      'foo/baz/foo.ts',
      'bar/foo/foo.ts',
      'bar/bar/foo.ts',
      'bar/baz/foo.ts',
      'baz/foo/foo.ts',
      'baz/bar/foo.ts',
      'baz/baz/foo.ts',
      'foo/foo/bar.ts',
      'foo/bar/bar.ts',
      'foo/baz/bar.ts',
      'bar/foo/bar.ts',
      'bar/bar/bar.ts',
      'bar/baz/bar.ts',
      'baz/foo/bar.ts',
      'baz/bar/bar.ts',
      'baz/baz/bar.ts',
    ])
  })

  it('excludes test files by default', () => {
    const root = fixtureRoot('test-files')
    const { findUncommented } = createFinder(root)

    const nodes = findUncommented()

    expectFiles(nodes, ['foo.ts'])
    expectNoFiles(nodes, ['foo.test.ts'])
  })

  it('excludes d.ts files', () => {
    const root = fixtureRoot('dts')
    const { findUncommented } = createFinder(root)

    const nodes = findUncommented()

    expectFiles(nodes, ['foo.ts'])
    expectNoFiles(nodes, ['foo.d.ts'])
  })

  it('supports glob exclude patterns', () => {
    const root = fixtureRoot('tree')
    const { findUncommented } = createFinder(root, {
      exclude: ['**/{foo,baz}.ts'],
    })

    const nodes = findUncommented()

    expectFiles(nodes, [
      'foo/foo/bar.ts',
      'foo/bar/bar.ts',
      'foo/baz/bar.ts',
      'bar/foo/bar.ts',
      'bar/bar/bar.ts',
      'bar/baz/bar.ts',
      'baz/foo/bar.ts',
      'baz/bar/bar.ts',
      'baz/baz/bar.ts',
    ])
  })

  it('finds interface properties', () => {
    const root = fixtureRoot('iface')
    const { findUncommented } = createFinder(root, {
      include: ['foo.ts'],
    })

    const nodes = findUncommented()

    expectFindings(nodes, {
      'foo.ts': [
        'iface:Foo',
        'method:Foo.foo',
        'prop:Foo.bar',
        'prop:Foo.foobar',
      ],
    })
  })

  it('finds class properties', () => {
    const root = fixtureRoot('iface')
    const { findUncommented } = createFinder(root, {
      include: ['bar.ts'],
    })

    const nodes = findUncommented()

    expectFindings(nodes, {
      'bar.ts': ['class:Foo', 'prop:Foo.foo', 'prop:Foo.foobar'],
    })
  })

  describe(`'symbols' option`, () => {
    const root = fixtureRoot('exclude-symbols')

    const all = [
      'iface:Foo',
      'method:Foo.foo',
      'class:Bar',
      'method:Bar.foo',
      'var:foobar',
      'func:foo',
      'var:bar',
    ] as const

    const tests = [
      {
        symbols: symbolTypes,
        want: all,
      },
      {
        symbols: [],
        want: all,
      },
      {
        symbols: without(['class'], symbolTypes),
        want: without(['class:Bar'], all),
      },
      {
        symbols: without(['var'], symbolTypes),
        want: without(['var:foobar', 'var:bar'], all),
      },
      {
        symbols: without(['func'], symbolTypes),
        want: without(['func:foo'], all),
      },
      {
        symbols: without(['method'], symbolTypes),
        want: without(['method:Foo.foo', 'method:Bar.foo'], all),
      },
      {
        symbols: without(['iface'], symbolTypes),
        want: without(['iface:Foo'], all),
      },
      {
        symbols: without(['class', 'var'], symbolTypes),
        want: without(['class:Bar', 'var:foobar', 'var:bar'], all),
      },
      {
        symbols: without(['class', 'var', 'func'], symbolTypes),
        want: without(['class:Bar', 'var:foobar', 'var:bar', 'func:foo'], all),
      },
      {
        symbols: without(['class', 'var', 'func', 'method'], symbolTypes),
        want: without(
          [
            'class:Bar',
            'var:foobar',
            'var:bar',
            'func:foo',
            'method:Foo.foo',
            'method:Bar.foo',
          ],
          all,
        ),
      },
    ] as const

    for (const [i, tt] of tests.entries()) {
      it(`allows configuring the symbols types to search for (#${i})`, () => {
        const { findUncommented } = createFinder(root, { symbols: tt.symbols })
        const nodes = findUncommented()

        expectFindings(nodes, { 'foo.ts': tt.want })
      })
    }
  })

  it('works with JS code', () => {
    const root = fixtureRoot('js')
    const { findUncommented } = createFinder(root)

    const nodes = findUncommented()

    expectFindings(nodes, {
      'foo.mjs': ['func:foo', 'class:Foo', 'prop:Foo.foo', 'method:Foo.foobar'],
    })
  })
})

describe('printFindings', () => {
  it('prints findings', () => {
    const root = fixtureRoot('iface')
    const { findUncommented } = createFinder(root, {
      include: ['foo.ts', 'bar.ts'],
    })

    const findings = findUncommented()

    const withoutNodes = removeNodesFromFindings(findings)
    const parsedFindings = JSON.parse(printFindings(findings)) as unknown

    expect(isEqual(parsedFindings, withoutNodes)).toBe(true)
  })
})
