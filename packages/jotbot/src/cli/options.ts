import type { Command } from 'commander'

/**
 * The `WithVerboseOption` interface provides an optional `verbose` property
 * that, when set to true, enables verbose output for the associated operations
 * or commands. This option is typically used to provide additional details or
 * debugging information during execution.
 */
export interface WithVerboseOption {
  verbose?: boolean
}

/**
 * `WithSourceOption` is an interface representing a configuration object that
 * optionally includes a path. When present, the path property specifies the
 * location of a resource or file to be used by the implementing entity. This
 * interface is useful for defining command-line options or configuration
 * settings where the source path may be provided by the user or inferred by the
 * context.
 */
export interface WithSourceOption {
  path?: string
}

/**
 * The `WithFormatOption` interface defines configuration options for output
 * formatting. It includes an optional `format` property, which specifies the
 * desired output format and must be one of the supported string literals.
 * Additionally, if 'json' is among the supported formats, a `json` property is
 * available as a boolean flag to indicate preference for JSON output. This
 * interface allows for consistent handling of format options across different
 * commands or functions that require output customization. When used in
 * conjunction with parsing or applying these options, a {@link Command} can
 * provide users with the ability to specify their desired format, potentially
 * defaulting to a predefined value if none is specified.
 */
export interface WithFormatOption<Supported extends string> {
  format?: Supported
  json?: 'json' extends Supported ? boolean : never
}

/**
 * Indicates whether to include additional information in the output, enabling
 * more detailed logging for diagnostic purposes. This tuple can be used as a
 * parameter to configure command-line interface options, where `-v` and
 * `--verbose` are the flags and 'Verbose output' is the description.
 */
export const verboseOption = ['-v, --verbose', 'Verbose output'] as const

/**
 * Configures command-line interface (CLI) options related to output formatting,
 * including the addition of a JSON option if supported. It applies default
 * formatting options and provides help text detailing supported formats. The
 * function takes an array of supported format strings, with an optional
 * configuration object that can specify a default format. Returns a function
 * that modifies a {@link Command} instance with the specified formatting
 * options.
 */
export function formatOptions<const Supported extends readonly string[]>(
  supported: Supported,
  options?: {
    default?: Supported[number]
  },
) {
  return (cmd: Command) => {
    cmd.option(...formatOption(supported, options))

    if (supported.includes('json')) {
      cmd.option(
        '--json',
        'Output findings as JSON (same as `--format json`)',
        false,
      )
    }

    cmd.addHelpText(
      'after',
      `\nSupported formats:\n  ${supported
        .map((format) => `"${format}"`)
        .join(', ')}`,
    )

    return cmd
  }
}

/**
 * Formats the given options for output, ensuring the format specified is one of
 * the supported types. If a default format is provided in the options, it is
 * used when no other format is specified. Throws an error if an unsupported
 * format is given. Returns an array containing the command line option
 * signature for specifying the output format, a description string (which
 * includes the default format if one is provided), a validation function, and
 * the default format value, if any. The returned array is intended for use with
 * a {@link Command} instance to configure its supported output formats.
 */
export function formatOption<const Supported extends readonly string[]>(
  supported: Supported,
  options?: {
    default?: Supported[number]
  },
) {
  return [
    '-f, --format <format>',
    `Configure formatting ${
      options?.default ? `(default: "${options.default}")` : ''
    }}`,
    (format: string) => {
      if (!supported.includes(format)) {
        throw new Error(`Unsupported format: ${format}`)
      }
      return format
    },
    options?.default,
  ] as const
}

/**
 * Parses the provided formatting options to determine the output format. It
 * prioritizes the JSON format if the `json` option is set to true within the
 * given {@link WithFormatOption}. If no specific format is selected, it falls
 * back to a provided default. The function ensures that the output format is
 * one of the supported formats, potentially including 'json'. It returns either
 * 'json' or another supported format as specified by the input constraints.
 */
export function parseFormat<Supported extends string>(
  options: WithFormatOption<Supported>,
  _default: Supported,
): 'json' extends Supported ? 'json' | Supported : Supported {
  return (options.json ? 'json' : options.format ?? _default) as any
}
