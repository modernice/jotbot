export interface Foo {
  foo(): string
}

export class Bar {
  foo(): string {
    return ''
  }
}

export const foobar = 'foobar'

export function foo(): string {
  return 'foo'
}

export const bar = () => 'bar' as const
