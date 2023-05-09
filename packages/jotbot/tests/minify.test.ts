import { describe, expect, it } from 'vitest'
import { heredoc } from '../src/utils'
import {
  MinifyFlags,
  defaultMinificationSteps,
  minify,
  minifyTo,
} from '../src/minify'
import { encoding_for_model } from '@dqbd/tiktoken'

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

      const { minified } = minify(code)

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

      const { minified } = minify(code)

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

      const { minified } = minify(code)

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

      const { minified } = minify(code)

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

      const { minified } = minify(code)

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

      const { minified } = minify(code)

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

      const { minified } = minify(code, { variables: false, emptyLines: false })

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

      const { minified } = minify(code, { functions: false, emptyLines: false })

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

      const { minified } = minify(code, { classes: false, emptyLines: false })

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

      const { minified } = minify(code, {
        interfaces: false,
        emptyLines: false,
      })

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

      const { minified } = minify(code, {
        properties: false,
        emptyLines: false,
      })

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

      const { minified } = minify(code, { methods: false, emptyLines: false })

      expect(minified).toEqual(code)
    })
  })

  describe('with token limit', () => {
    it("doesn't compute tokens by default", () => {
      const code = heredoc`
				/**
				 * foo is a function.
				 */
				function foo() {}
			`

      // @ts-expect-error
      const { inputTokens, tokens } = minify(code)

      expect(inputTokens).toBeUndefined()
      expect(tokens).toBeUndefined()
    })

    it('returns the tokens of the input code and minified code', () => {
      const code = heredoc`
				/**
				 * foo is a function.
				 */
				function foo() {}
			` // 15 tokens

      const wantMinified = heredoc`
				function foo() {}
			` // 4 tokens

      const { tokens, inputTokens, minified } = minify(code, {
        model: 'text-davinci-003',
        computeTokens: true,
      })

      expect(minified).toEqual(wantMinified)
      expect(inputTokens).toHaveLength(15)
      expect(tokens).toHaveLength(4)
    })
  })
})

describe('minifyTo', () => {
  it("doesn't minify if input doesn't hit maxTokens limit", () => {
    const code = heredoc`
			/**
			 * foo is a function.
			 */
			function foo() {}
		` // 15 tokens

    const { tokens, inputTokens, minified, steps } = minifyTo(15, code, {
      model: 'text-davinci-003',
    })

    expect(minified).toEqual(code)
    expect(inputTokens).toHaveLength(15)
    expect(tokens).toHaveLength(15)
    expect(steps).toHaveLength(0)
  })

  it('minifies if input hits maxTokens limit', () => {
    const code = heredoc`
			/**
			 * foo is a function.
			 */
			function foo() {}
		` // 15 tokens

    const wantMinified = heredoc`
			function foo() {}
		` // 4 tokens

    const { tokens, inputTokens, minified, steps } = minifyTo(14, code, {
      model: 'text-davinci-003',
    })

    expect(minified).toEqual(wantMinified)
    expect(inputTokens).toHaveLength(15)
    expect(tokens).toHaveLength(4)
    expect(steps).toHaveLength(3)

    expect(steps[0]).toEqual({
      flags: minifyFlags(defaultMinificationSteps[0]),
      input: code,
      minified: code,
      inputTokens: computeTokens(code),
      tokens: computeTokens(code),
    })

    expect(steps[1]).toEqual({
      flags: minifyFlags(defaultMinificationSteps[1]),
      input: code,
      minified: code,
      inputTokens: computeTokens(code),
      tokens: computeTokens(code),
    })

    expect(steps[2]).toEqual({
      flags: minifyFlags(defaultMinificationSteps[2]),
      input: code,
      minified: wantMinified,
      inputTokens: computeTokens(code),
      tokens: computeTokens(wantMinified),
    })
  })

  it('accepts custom minification steps', () => {
    const code = heredoc`
			/**
			 * foo is a function.
			 */
			function foo() {}
		` // 15 tokens

    const wantMinified = heredoc`
			function foo() {}
		` // 4 tokens

    const { tokens, inputTokens, minified, steps } = minifyTo(14, code, {
      model: 'text-davinci-003',
      steps: [{ functions: true }],
    })

    expect(minified).toEqual(wantMinified)
    expect(inputTokens).toHaveLength(15)
    expect(tokens).toHaveLength(4)
    expect(steps).toHaveLength(1)

    expect(steps[0]).toEqual({
      flags: minifyFlags({ functions: true }),
      input: code,
      minified: wantMinified,
      inputTokens: computeTokens(code),
      tokens: computeTokens(wantMinified),
    })
  })
})

function minifyFlags(flags: Partial<MinifyFlags>): MinifyFlags {
  return {
    variables: flags.variables ?? false,
    functions: flags.functions ?? false,
    classes: flags.classes ?? false,
    interfaces: flags.interfaces ?? false,
    properties: flags.properties ?? false,
    methods: flags.methods ?? false,
    emptyLines: flags.emptyLines ?? false,
  }
}

const enc = encoding_for_model('text-davinci-003')
const tokenCache = new Map<string, Uint32Array>()
function computeTokens(code: string) {
  const cached = tokenCache.get(code)
  if (cached) {
    return cached
  }
  const tokens = enc.encode(code)
  tokenCache.set(code, tokens)
  return tokens
}
