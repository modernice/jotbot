import { resolve } from 'node:path'
import type { Command } from 'commander'
import type { FinderOptions, SymbolType } from '..'
import { createFinder, defaultExclude, isSymbol, printFindings } from '..'
import { createLogger } from './logger'
import type { WithFormatOptions, WithVerboseOption } from './options'
import { commaSeparated } from './utils'

const { log: print } = createLogger(process.stdout)

export function withFindCmd(program: Command) {
  program
    .command('find')
    .description('Find uncommented symbols in <root>')
    .argument('[root]', 'Root directory to search', '.')
    .option(
      '-s, --symbols <symbols>',
      'Symbols to search for (comma-separated)',
      commaSeparated({ validate: isSymbol }),
      [] as SymbolType[],
    )
    .option(
      '-i, --include <patterns>',
      'Include files matching glob patterns (comma-separated)',
      commaSeparated(),
    )
    .option(
      '-e, --exclude <patterns>',
      'Exclude files matching glob patterns (comma-separated)',
      commaSeparated(),
    )
    .option('-f, --format', 'Configure formatting (default: "list")', 'list')
    .option('--json', 'Output findings as JSON (overrides `--format`)', false)
    .option('-v, --verbose', 'Verbose output', false)
    .addHelpText(
      'after',
      '\nDefault --symbols:\n  ["func", "var", "class", "method", "iface", "prop"]',
    )
    .addHelpText('after', '\nDefault --include:\n  [**/*]')
    .addHelpText(
      'after',
      `\nDefault --exclude:\n  ${JSON.stringify(defaultExclude, null, 2)
        .split('\n')
        .join('\n  ')}`,
    )
    .addHelpText('after', '\nSupported formats:\n  ["list", "json"]')
    .action(run)

  return program
}

function run(
  root: string,
  options: FinderOptions & WithVerboseOption & WithFormatOptions,
) {
  const format = options.json ? 'json' : options.format ?? 'list'

  root = resolveRoot(root)

  const { log, info } = createLogger(process.stderr, {
    enabled: options.verbose ?? false,
  })

  info(`Root: ${root}`)
  info(`Symbols: ${options.symbols?.join(', ') || '-'}`)
  info(`Include: ${options.include ?? '-'}`)
  info(`Exclude: ${options.exclude ?? '-'}`)

  const { findUncommented } = createFinder(root, options)

  info(`Searching for uncommented symbols in ${root} ...\n`)

  const findings = findUncommented()
  const entries = Object.entries(findings)

  if (!entries.length) {
    log(`No uncommented symbols found in ${root}.`)
    return
  }

  if (format === 'list') {
    for (const [file, _findings] of Object.entries(findings)) {
      for (const finding of _findings) {
        print(`${file}@${finding.identifier}`)
      }
    }
    return
  }

  if (format === 'json') {
    print(printFindings(findings))
  }
}

function resolveRoot(root: string) {
  const workingDir = process.cwd()
  return resolve(workingDir, root)
}
