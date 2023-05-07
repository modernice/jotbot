import type { Command } from 'commander'

/**
 * The `iface:WithVerboseOption` is an interface representing an object with an
 * optional `verbose` property. The `verbose` property, when set to `true`,
 * indicates that verbose output should be enabled. This interface is typically
 * used in conjunction with command line utilities or tools to provide a more
 * detailed output for debugging or informational purposes.
 */
export interface WithVerboseOption {
  /**
   * Property "WithVerboseOption.verbose": Indicates whether the verbose output
   * option is enabled or not. When enabled, it provides more detailed
   * information about the operation being performed. This property is optional
   * and has a boolean value.
   */
  verbose?: boolean
}

/**
 * The `WithSourceOption` interface provides an optional `path` property, which
 * represents a string containing the source path for a specific operation. This
 * interface is used to define the source path option for commands and
 * configurations in the application.
 */
export interface WithSourceOption {
  /**
   * The "WithSourceOption.path" property represents an optional path to a
   * source file or directory. When specified, it provides the location of the
   * source data to be processed by the associated command. If not specified,
   * the command will use a default or fallback behavior to determine the source
   * location.
   */
  path?: string
}

/**
 * The `WithFormatOption` interface provides a way to specify the format of the
 * output. It includes an optional `format` property, which can be set to one of
 * the supported format strings, and an optional `json` property, which can be
 * set to a boolean value when the supported formats include 'json'. The `json`
 * property acts as a shortcut for setting the output format to 'json'.
 */
export interface WithFormatOption<Supported extends string> {
  /**
   * The `WithFormatOption.format` property is an optional configuration for
   * specifying the output format of the data. It accepts a supported format
   * string and allows the user to customize the display of the output. The
   * available formats are defined by the `Supported` type parameter. If a
   * default format is provided, it will be used when no format is specified by
   * the user. To reference this property, use {@link WithFormatOption.format}.
   */
  format?: Supported
  /**
   * The `WithFormatOption.json` property is a boolean flag that, when set to
   * `true`, indicates that the output should be in JSON format. This property
   * is only available when the supported formats include 'json'. It provides an
   * alternative way to specify JSON output, equivalent to using `--format
   * json`.
   */
  json?: 'json' extends Supported ? boolean : never
}

/**
 * The variable "verboseOption" is an array that represents a command-line
 * option for enabling verbose output. When the "-v" or "--verbose" flag is
 * provided, the output will include additional details and information.
 */
export const verboseOption = ['-v, --verbose', 'Verbose output'] as const

/**
 * The `formatOptions()` function is a utility that configures the
 * format-related options for a given command. It accepts a list of supported
 * formats and an optional configuration object with a default format. The
 * function adds the `--format` option to the command, allowing users to specify
 * the desired output format. If 'json' is included in the supported formats, it
 * also adds the `--json` option as an alternative way to set the output format
 * to JSON. Additionally, it appends a help text listing all supported formats.
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
 * The `formatOption()` function is used to configure the formatting option for
 * a command-line application. It takes two arguments: an array of supported
 * formats (supported) and an optional configuration object (options). The
 * configuration object can have a `default` property, which sets the default
 * format when none is specified by the user. The function returns an array
 * containing the format option flag, description, validation function, and
 * default value. If an unsupported format is provided by the user, the
 * validation function throws an error.
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
 * The `parseFormat()` function is used to determine the output format based on
 * the provided options. It takes two arguments: `options`, which is an object
 * with the format and json properties, and `_default`, which is a default
 * format value. The function returns the selected format as a string. If the
 * 'json' option is supported and set to true, the function will return 'json'
 * as the format. Otherwise, it will return either the specified format in the
 * options object or the default format provided.
 */
export function parseFormat<Supported extends string>(
  options: WithFormatOption<Supported>,
  _default: Supported,
): 'json' extends Supported ? 'json' | Supported : Supported {
  return (options.json ? 'json' : options.format ?? _default) as any
}
