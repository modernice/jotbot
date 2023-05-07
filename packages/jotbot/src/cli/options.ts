import type { Command } from 'commander'

/**
 * Specifies an optional `verbose` flag in an object that can be used to enable
 * verbose output.
 */
export interface WithVerboseOption {
  /** Enables verbose output when set to true. */
  verbose?: boolean
}

/**
 * Specifies an interface for options that include a file path to the source
 * code being analyzed.
 */
export interface WithSourceOption {
  /**
   * Specifies the path to the file or directory to be processed, as an optional
   * property of an object.
   */
  path?: string
}

/**
 * Specifies an interface for options related to formatting, including the
 * ability to set the output format and output findings in JSON format if
 * supported.
 */
export interface WithFormatOption<Supported extends string> {
  /** Specifies the format of the output. */
  format?: Supported
  /**
   * Specifies an optional format for output, with support for JSON format, as
   * well as the ability to configure the format.
   */
  json?: 'json' extends Supported ? boolean : never
}

/**
 * The `verboseOption` variable represents the command line option for enabling
 * verbose output.
 */
export const verboseOption = ['-v, --verbose', 'Verbose output'] as const

/**
 * Formats the options for a command by adding a format option and a json option
 * if 'json' is included in the supported formats, and adds help text displaying
 * the supported formats. Returns the modified command with the added options.
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
 * Formats the output of a command by configuring the format option and
 * validating the supported formats.
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
 * Parses the format option from the provided options object or returns the
 * default format. If the JSON option is set to true, returns 'json' if it is a
 * supported format, otherwise throws an error.
 */
export function parseFormat<Supported extends string>(
  options: WithFormatOption<Supported>,
  _default: Supported,
): 'json' extends Supported ? 'json' | Supported : Supported {
  return (options.json ? 'json' : options.format ?? _default) as any
}
