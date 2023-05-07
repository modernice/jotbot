/**
 * A function that takes an optional object with a validation function and a
 * parsing function. It splits a string value by commas and validates each value
 * if a validation function is provided. If a parsing function is provided, it
 * will parse each value accordingly. The function returns an array of parsed
 * values concatenated with the previous array of values passed as a second
 * argument.
 */
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
