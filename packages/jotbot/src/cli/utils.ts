/**
 * Transforms a comma-separated string into an array of type T, optionally
 * performing validation and custom parsing on each separated value. If
 * validation is provided and fails for any value, an error is thrown. The
 * function returns an array that combines the previously accumulated values
 * with the newly parsed ones, resulting in an aggregated array of type {@link
 * T}.
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
