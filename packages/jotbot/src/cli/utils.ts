export function commaSeparated<T = string>(
  options?: { validate?: (value: string) => boolean } & (T extends string
    ? { parse?: (value: string) => T }
    : { parse: (value: string) => T }),
) {
  return (value: string, prev?: readonly T[]) => {
    const values = value.split(',')

    if (
      options?.validate &&
      values.some((value) => options.validate && !options.validate(value))
    ) {
      throw new Error(`invalid option value: ${values.join(', ')}`)
    }

    const parsed = options?.parse
      ? values.map((value) => options?.parse?.(value) ?? value)
      : values

    return (prev ?? []).concat(parsed as T[])
  }
}
