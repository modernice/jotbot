import { describe, expect, it } from 'vitest'
import { findClass, findFunction, findMethod, findVariable } from '../src/nodes'
import { formatComment, updateNodeComments } from '../src/comments'
import { heredoc } from '../src/utils'
import { parseCode } from '../src/parse'
import { printFile } from '../src/print'
import { withFixture } from './testutils'

describe('updateNodeComments', () => {
  it("puts the comment before an 'export const' variable declaration", () => {
    const code = heredoc`
      export const foo = 'bar';
    `

    const file = parseCode(code)
    const result = findVariable(file, 'foo')

    expect(result).toBeTruthy()

    const { commentTarget } = result!

    updateNodeComments(commentTarget, [formatComment("foo is always 'bar'.")])

    const text = printFile(file)

    expect(text).toBe(heredoc.withNewline`
      /**
       * foo is always 'bar'.
       */
      export const foo = 'bar';
    `)
  })

  it("puts the comment before the 'export function' declaration", () => {
    const code = heredoc`
      export function foo() {}
    `

    const file = parseCode(code)
    const declaration = findFunction(file, 'foo')

    expect(declaration).toBeTruthy()

    const { commentTarget } = declaration!

    updateNodeComments(commentTarget, [formatComment('foo is a function.')])

    const text = printFile(file)

    expect(text).toBe(heredoc.withNewline`
      /**
       * foo is a function.
       */
      export function foo() { }
    `)
  })

  it("puts the comment before the 'export class' declaration", () => {
    const code = heredoc`
      export class Foo {}
    `

    const file = parseCode(code)
    const declaration = findClass(file, 'Foo')

    expect(declaration).toBeTruthy()

    const { commentTarget } = declaration!

    updateNodeComments(commentTarget, [formatComment('Foo is a class.')])

    const text = printFile(file)

    expect(text).toBe(heredoc.withNewline`
      /**
       * Foo is a class.
       */
      export class Foo {
      }
    `)
  })

  it('puts the comment before the class method', () => {
    const code = heredoc`
      export class Foo {
        foo() {}
      }
    `

    const file = parseCode(code)
    const method = findMethod(file, 'Foo', 'foo')

    expect(method).toBeTruthy()

    const { commentTarget } = method!

    updateNodeComments(commentTarget, [formatComment('foo is a method.')])

    const text = printFile(file)

    expect(text).toBe(heredoc.withNewline`
      export class Foo {
          /**
           * foo is a method.
           */
          foo() { }
      }
    `)
  })
})

describe('applyComments', () => {
  it('applies variable comments', () => {
    withFixture('basic', ({ testPatch }) => {
      testPatch('foo.ts', {
        'var:foobar': {
          comment: 'foobar is a const string.',
          want: heredoc`
            /**
             * foobar is a const string.
             */
          `,
        },
      })
    })
  })

  it('applies function comments', () => {
    withFixture('basic', ({ testPatch }) => {
      testPatch('foo.ts', {
        'func:foo': {
          comment: 'foo is a function that returns "foo".',
          want: heredoc`
            /**
             * foo is a function that returns "foo".
             */
          `,
        },
      })
    })
  })

  it('applies class comments', () => {
    withFixture('basic', ({ testPatch }) => {
      testPatch('bar.ts', {
        'class:Bar': {
          comment: 'Bar is a class that has a method bar.',
          want: heredoc`
            /**
             * Bar is a class that has a method bar.
             */
          `,
        },
        'method:Bar.bar': {
          comment: 'bar is a method that returns "bar".',
          want: heredoc`
            /**
             * bar is a method that returns "bar".
             */
          `,
        },
      })
    })
  })

  it('applies interface comments', () => {
    withFixture('iface', ({ testPatch }) => {
      testPatch('foo.ts', {
        'iface:Foo': {
          comment: 'Foo is an interface.',
          want: heredoc`
            /**
             * Foo is an interface.
             */
          `,
        },
        'method:Foo.foo': {
          comment: 'foo is a method that returns a string.',
          want: heredoc`
            /**
             * foo is a method that returns a string.
             */
          `,
        },
      })
    })
  })
})
