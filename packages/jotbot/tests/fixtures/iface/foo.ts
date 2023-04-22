export interface Foo {
  readonly foobar: string

  foo(): string
  bar: () => string
}

interface Bar {
  foo(): string
  bar: () => string
}
