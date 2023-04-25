import { describe, expect, it } from 'vitest'
import {
  NaturalLanguageTarget,
  RawIdentifier,
  describeIdentifier,
} from '../src'

describe('describeIdentifier', () => {
  const tests = {
    'var:foo': `variable 'foo'`,
    'func:foo': `function 'foo'`,
    'class:Foo': `class 'Foo'`,
    'iface:Foo': `interface 'Foo'`,
    'method:Foo.foo': `method 'foo' of 'Foo'`,
    'prop:Foo.foo': `property 'foo' of 'Foo'`,
  } as const satisfies Record<RawIdentifier, NaturalLanguageTarget>

  for (const [identifier, want] of Object.entries(tests)) {
    it(`creates a natural language target for '${identifier}'`, () => {
      const target = describeIdentifier(identifier as RawIdentifier)
      expect(target).toBe(want)
    })
  }
})
