import { describe, expect, it } from 'vitest'
import { heredoc } from '../src/utils'
import { minify } from '../src/minify'

describe('minify', () => {
  describe('by default', () => {
    it('removes variable comments', () => {
      const code = heredoc`
				/**
				 * foo is a variable.
				 */
				const foo = 'foo'

				/**
				 * bar is a variable.
				 */
				export const bar = 'bar'
			`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				const foo = 'foo'
				export const bar = 'bar'
			`)
    })

    it('removes function comments', () => {
      const code = heredoc`
				/**
				 * foo is a function.
				 */
				function foo() {
					return 'foo'
				}

				/**
				 * bar is a function.
				 */
				export function bar() {
					return 'bar'
				}
			`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				function foo() {
					return 'foo'
				}
				export function bar() {
					return 'bar'
				}
			`)
    })

    it('removes class comments', () => {
      const code = heredoc`
			/**
			 * Foo is a class.
			 */
			class Foo {}

			/**
			 * Bar is a class.
			 */
			export class Bar {}
		`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				class Foo {}
				export class Bar {}
			`)
    })

    it('removes interface comments', () => {
      const code = heredoc`
				/**
				 * Foo is an interface.
				 */
				interface Foo {}

				/**
				 * Bar is an interface.
				 */
				export interface Bar {}
			`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				interface Foo {}
				export interface Bar {}
			`)
    })

    it('removes property comments', () => {
      const code = heredoc`
				interface Foo {
					/**
					 * foo is a property.
					 */
					foo: string
				}

				class Bar {
					/**
					 * bar is a property.
					 */
					bar: string
				}
			`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				interface Foo {
					foo: string
				}
				class Bar {
					bar: string
				}
			`)
    })

    it('removes method comments', () => {
      const code = heredoc`
				interface Foo {
					/**
					 * foo is a method.
					 */
					foo(): string
				}

				class Bar {
					/**
					 * bar is a method.
					 */
					bar(): string
				}
			`

      const minified = minify(code)

      expect(minified).toEqual(heredoc`
				interface Foo {
					foo(): string
				}
				class Bar {
					bar(): string
				}
			`)
    })
  })

  describe('custom options', () => {
    it("doesn't remove variable comments if options.variables is false", () => {
      const code = heredoc`
  			/**
  			 * foo is a variable.
  			 */
  			const foo = 'foo'

  			/**
  			 * bar is a variable.
  			 */
  			export const bar = 'bar'
  		`

      const minified = minify(code, { variables: false, emptyLines: false })

      expect(minified).toEqual(code)
    })

    it("doesn't remove function comments if options.functions is false", () => {
      const code = heredoc`
				/**
				 * foo is a function.
				 */
				function foo() {
					return 'foo'
				}

				/**
				 * bar is a function.
				 */
				export function bar() {
					return 'bar'
				}
			`

      const minified = minify(code, { functions: false, emptyLines: false })

      expect(minified).toEqual(code)
    })

    it("doesn't remove class comments if options.classes is false", () => {
      const code = heredoc`
				/**
				 * Foo is a class.
				 */
				class Foo {}

				/**
				 * Bar is a class.
				 */
				export class Bar {}
			`

      const minified = minify(code, { classes: false, emptyLines: false })

      expect(minified).toEqual(code)
    })

    it("doesn't remove interface comments if options.interfaces is false", () => {
      const code = heredoc`
				/**
				 * Foo is an interface.
				 */
				interface Foo {}

				/**
				 * Bar is an interface.
				 */
				export interface Bar {}
			`

      const minified = minify(code, { interfaces: false, emptyLines: false })

      expect(minified).toEqual(code)
    })

    it("doesn't remove property comments if options.properties is false", () => {
      const code = heredoc`
				interface Foo {
					/**
					 * foo is a property.
					 */
					foo: string
				}

				class Bar {
					/**
					 * bar is a property.
					 */
					bar: string
				}
			`

      const minified = minify(code, { properties: false, emptyLines: false })

      expect(minified).toEqual(code)
    })

    it("doesn't remove method comments if options.methods is false", () => {
      const code = heredoc`
				interface Foo {
					/**
					 * foo is a method.
					 */
					foo(): string
				}

				class Bar {
					/**
					 * bar is a method.
					 */
					bar(): string
				}
			`

      const minified = minify(code, { methods: false, emptyLines: false })

      expect(minified).toEqual(code)
    })
  })
})
