import { describe, expect, it } from 'vitest'
import { heredoc } from '../src/utils'

describe('heredoc', () => {
  it('works', () => {
    const result = heredoc`
      foo
      bar
    `

    expect(result).toBe('foo\nbar')
  })

  it('works with indentation', () => {
    const result = heredoc`
      foo
        bar
      foobar
          baz
    `

    expect(result).toBe(`foo\n${tab()}bar\nfoobar\n${tab().repeat(2)}baz`)
  })

  it('works with mixed indentation', () => {
    const result = heredoc`
      foo
        bar
      foobar
          baz
    `

    expect(result).toBe(`foo\n${tab()}bar\nfoobar\n${tab().repeat(2)}baz`)
  })

  it('works with multi-line trivia comments', () => {
    const result = heredoc`
      /**
       * foo is a constant string.
       */
      const foo = 'foo'
    `

    expect(result).toBe(
      '/**\n * foo is a constant string.\n */\nconst foo = \'foo\'',
    )
  })
})

describe('heredoc.withNewline', () => {
  it('adds a newline after the result', () => {
    const input = `
      foo
        bar
          baz
        foobar
    `

    const want = `${heredoc`${input}`}\n`
    const got = heredoc.withNewline`${input}`

    expect(got).toBe(want)
  })
})

function tab(width = 2) {
  return ' '.repeat(width)
}
