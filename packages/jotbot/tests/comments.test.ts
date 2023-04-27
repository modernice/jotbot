import { describe, expect, it } from 'vitest'
import { findClass, findFunction, findMethod, findVariable } from '../src/nodes'
import { formatComment, updateNodeComments } from '../src/comments'
import { heredoc } from '../src/utils'
import { parseCode } from '../src/parse'
import { printComment } from '../src'

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

    expect(printComment(commentTarget)).toBe(heredoc`
      /**
       * foo is always 'bar'.
       */
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

    expect(printComment(commentTarget)).toBe(heredoc`
      /**
       * foo is a function.
       */
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

    expect(printComment(commentTarget)).toBe(heredoc`
      /**
       * Foo is a class.
       */
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

    expect(printComment(commentTarget)).toBe(heredoc`
      /**
       * foo is a method.
       */
    `)
  })
})
