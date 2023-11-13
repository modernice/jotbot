/**
 * Converts a single item or an array of items into an array of elements. If no
 * value is provided, it returns an empty array. When a single item is given, it
 * wraps the item in an array. If an array is passed, it simply returns the same
 * array.
 */
export function toArray<T>(value?: T | readonly T[]): T[] {
  if (!value) {
    return []
  }
  return Array.isArray(value) ? (value as any) : [value]
}

/**
 * Processes a template string with embedded expressions into a single formatted
 * string. The function formats multi-line string literals, known as heredoc, by
 * removing excessive indentation and ensuring consistent whitespace handling.
 * It interpolates given values into the template strings and aligns the
 * resulting text to the left-most non-whitespace character, trimming any
 * leading and trailing whitespace from the final output. Additionally, it
 * provides a `withNewline` variant that appends a newline character to the end
 * of the processed string. Returns a {@link String} representing the formatted
 * heredoc output.
 */
export function heredoc(
  strings: TemplateStringsArray,
  ...values: any[]
): string {
  let result = ''

  for (let i = 0; i < strings.length; i++) {
    result += strings[i]
    if (i < values.length) {
      result += values[i]
    }
  }

  result = result.replace(/^\s+/gm, (match) => {
    return match.replace(/\t/g, '  ')
  })

  const lines = result.split('\n')
  const minIndent = Math.min(
    ...lines
      .filter((line) => line.trim().length > 0)
      .map((line) => line.match(/^\s*/)?.[0]?.length || 0),
  )
  const strippedLines = lines.map((line) => line.slice(minIndent))

  return strippedLines.join('\n').trim()
}

heredoc.withNewline = (strings: TemplateStringsArray, ...values: any[]) => {
  const vals = strings.reduce<string[]>((acc, val, i) => {
    return [...acc, val, String(values?.[i] ?? '')]
  }, [])

  return `${heredoc`${vals.join('')}`}\n`
}
