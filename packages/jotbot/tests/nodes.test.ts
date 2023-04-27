import { describe, expect, it } from 'vitest'
import { findClass, findFunction, findMethod, findVariable } from '../src/nodes'
import { heredoc } from '../src/utils'
import { parseCode } from '../src/parse'
import { printSourceFileComments, printSyntheticComments } from '../src/print'
import { patchComments } from '../src'

describe('findVariable', () => {
  it("doesn't find unexported variables", () => {
    const code = heredoc`
      const foo = 'foo';
      var bar = 'bar';
      const baz = () => {};
    `

    const file = parseCode(code)

    for (const name of ['foo', 'bar', 'baz']) {
      const result = findVariable(file, name)
      expect(result).toBeNull()
    }
  })

  it('finds exported variables', () => {
    const code = heredoc`
      export const foo = 'foo';
      export var bar = 'bar';
      export const baz = () => {};
      let foobar = 'foobar';
    `

    const file = parseCode(code)

    const result = findVariable(file, 'foobar')
    expect(result).toBeNull()

    for (const name of ['foo', 'bar', 'baz']) {
      const result = findVariable(file, name)
      expect(result).toBeTruthy()
    }
  })

  it("doesn't find nested variables within exported functions", () => {
    const code = heredoc`
      export function foo() {
        const bar = 'bar';
      }
    `

    const file = parseCode(code)

    expect(findVariable(file, 'foo')).toBeNull() // not a variable
    expect(findFunction(file, 'foo')).toBeTruthy()
    expect(findVariable(file, 'bar')).toBeNull() // not exported
  })

  it("doesn't find nested functions within exported functions", () => {
    const code = heredoc`
      export function foo() {
        function bar() {}
      }
    `

    const file = parseCode(code)

    expect(findFunction(file, 'foo')).toBeTruthy()
    expect(findVariable(file, 'bar')).toBeNull() // not exported
  })
})

describe('findFunction', () => {
  it("doesn't find unexported functions", () => {
    const code = heredoc`
      function foo() {}
      function bar() {}
    `

    const file = parseCode(code)

    for (const name of ['foo', 'bar']) {
      const result = findFunction(file, name)
      expect(result).toBeNull()
    }
  })

  it('finds exported functions', () => {
    const code = heredoc`
      export function foo() {}
      export function bar() {};
      export const baz = () => {}
    `

    const file = parseCode(code)

    for (const name of ['foo', 'bar']) {
      const result = findFunction(file, name)
      expect(result).toBeTruthy()
    }

    expect(findFunction(file, 'baz')).toBeNull() // baz is a variable
  })
})

describe('findClass', () => {
  it("doesn't find unexported classes", () => {
    const code = heredoc`
      class Foo {}
      const Bar = class {};
    `

    const file = parseCode(code)

    for (const name of ['Foo', 'Bar']) {
      const result = findClass(file, name)
      expect(result).toBeNull()
    }
  })

  it('finds exported classes', () => {
    const code = heredoc`
      export class Foo {}
      export const Bar = class {};
    `

    const file = parseCode(code)

    for (const name of ['Foo', 'Bar']) {
      const result = findClass(file, name)
      expect(result).toBeTruthy()
    }
  })

  it('doesnt find nested classes within exported functions', () => {
    const code = heredoc`
      export function foo() {
        class Bar {}
        const Baz = class {}
      }
    `

    const file = parseCode(code)

    expect(findClass(file, 'foo')).toBeNull() // not a class
    expect(findClass(file, 'Bar')).toBeNull()
    expect(findClass(file, 'bar')).toBeNull() // not exported
  })
})

describe('findMethod', () => {
  it("doesn't find unexported methods", () => {
    const code = heredoc`
      class Foo {
        foo() {}
      }
    `

    const file = parseCode(code)

    const result = findMethod(file, 'Foo', 'foo')
    expect(result).toBeNull()
  })

  it('finds exported methods', () => {
    const code = heredoc`
      export class Foo {
        foo() {}
      }
    `

    const file = parseCode(code)

    const result = findMethod(file, 'Foo', 'foo')
    expect(result).toBeTruthy()
  })

  it("doesn't find private methods", () => {
    const code = heredoc`
      export class Foo {
        private foo() {}
        bar() {}
      }
    `

    const file = parseCode(code)

    expect(findMethod(file, 'Foo', 'foo')).toBeNull()
    expect(findMethod(file, 'Foo', 'bar')).toBeTruthy()
  })
})

describe('printSourceFileComments', () => {
  it('prints source file comments', () => {
    const code = heredoc`
      /**
       * foo is a function.
       */
      export function foo() {}

      /**
       * Bar is a class.
       */
      export class Bar {}
    `

    const file = parseCode(code)
    const foo = findFunction(file, 'foo')!
    const bar = findClass(file, 'Bar')!

    const fooComment = printSourceFileComments(foo.commentTarget)
    const barComment = printSourceFileComments(bar.commentTarget)

    expect(fooComment).toBe(heredoc`
      /**
       * foo is a function.
       */
    `)

    expect(barComment).toBe(heredoc`
      /**
       * Bar is a class.
       */
    `)
  })
})

describe('printSyntheticComments', () => {
  it('prints synthetic comments', () => {
    const code = heredoc`
      export function foo() {}
      export class Bar {}
    `

    const file = parseCode(code)

    patchComments(file, {
      'func:foo': 'foo is a function.',
      'class:Bar': 'Bar is a class.',
    })

    const foo = findFunction(file, 'foo')!
    const bar = findClass(file, 'Bar')!

    const fooComment = printSyntheticComments(foo.commentTarget)
    const barComment = printSyntheticComments(bar.commentTarget)

    expect(fooComment).toBe(heredoc`
      /**
       * foo is a function.
       */
    `)
    expect(barComment).toBe(heredoc`
      /**
       * Bar is a class.
       */
    `)
  })
})
