/**
 * Converts the given value to an array. If the value is already an array, it
 * returns the value unchanged. If the value is undefined or null, it returns an
 * empty array. Otherwise, it returns a new array containing the value as its
 * single element.
 * 
 * @param value - The value to be converted to an array.
 * @returns An array containing the given value or an empty array if the value
 * is undefined or null.
 */
export function toArray<T>(value?: T | readonly T[]): T[] {
  if (!value) {
    return []
  }
  return Array.isArray(value) ? (value as any) : [value]
}

/**
 * The `heredoc()` function takes a template string with newline and indentation
 * characters, and returns a formatted string with consistent indentation. It
 * accepts a `TemplateStringsArray` and an optional array of values to be
 * interpolated within the template string. The function trims leading
 * whitespace characters and adjusts the indentation of each line based on the
 * minimum indentation found in the input string.
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
