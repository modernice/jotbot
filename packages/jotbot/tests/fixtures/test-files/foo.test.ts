import { describe, expect, it } from 'vitest'
import { foo } from './foo'

export function foobar() {}

describe('foo', () => {
  it(`returns 'foo'`, () => {
    expect(foo()).toBe('foo')
  })
})
