import type { Command } from 'commander'

export interface WithVerboseOption {
  verbose?: boolean
}

export interface WithSourceOption {
  path?: string
}

export interface WithFormatOption<Supported extends string> {
  format?: Supported
  json?: 'json' extends Supported ? boolean : never
}

export const verboseOption = ['-v, --verbose', 'Verbose output'] as const

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

export function parseFormat<Supported extends string>(
  options: WithFormatOption<Supported>,
  _default: Supported,
): 'json' extends Supported ? 'json' | Supported : Supported {
  return (options.json ? 'json' : options.format ?? _default) as any
}
