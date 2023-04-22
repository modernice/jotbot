export function toArray<T>(value?: T | readonly T[]): T[] {
  if (!value) {
    return []
  }
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  return Array.isArray(value) ? (value as any) : [value]
}

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
