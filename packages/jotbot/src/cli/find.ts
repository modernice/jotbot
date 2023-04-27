import type { Command } from 'commander'
import type { FinderOptions, SymbolType } from '..'
import {
  createFinder,
  isSymbol,
  printFindings,
  readSource,
  symbolTypes,
} from '..'
import { createLogger } from './logger'
import { commaSeparated } from './utils'

interface Options extends FinderOptions {
  path?: string
  verbose?: boolean
  format?: 'json' | 'list'
  json?: boolean
}

const { log: print } = createLogger(process.stdout)

export function withFindCmd(program: Command) {
  program
    .command('find')
    .description('Find uncommented symbols in TS/JS source code')
    .argument('[code]', 'TS/JS source code', '')
    .option('-p, --path <file>', 'Path to TS/JS file (instead of code)')
    .option(
      '-s, --symbols <symbols>',
      'Symbols to search for (comma-separated)',
      commaSeparated({ validate: isSymbol }),
      [] as SymbolType[],
    )
    .option('-f, --format', 'Configure formatting (default: "list")', 'list')
    .option(
      '--json',
      'Output findings as JSON (same as `--format json`)',
      false,
    )
    .option('-v, --verbose', 'Verbose output', false)
    .addHelpText(
      'after',
      '\nDefault --symbols:\n  ["func", "var", "class", "method", "iface", "prop"]',
    )
    .addHelpText('after', '\nSupported formats:\n  ["list", "json"]')
    .action(run)

  return program
}

function run(code: string, options: Options) {
  const format = options.json ? 'json' : options.format ?? 'list'

  const { log, info } = createLogger(process.stderr, {
    enabled: options.verbose ?? false,
  })

  if (options.path) {
    info(`File: ${options.path}`)
    code = readSource(options.path)
  }
  info(
    `Symbols: ${(options.symbols?.length ? options.symbols : symbolTypes).join(
      ', ',
    )}`,
  )

  const { find } = createFinder(options)

  if (options.path) {
    info(`Searching for uncommented symbols in ${options.path} ...`)
  } else {
    info(`Searching for uncommented symbols ...`)
  }

  const findings = find(code)
  const entries = Object.entries(findings)

  if (!entries.length) {
    log(`No uncommented symbols found.`)
  }

  if (format === 'list') {
    log('')
    for (const finding of findings) {
      print(finding.identifier)
    }
    return
  }

  if (format === 'json') {
    print(printFindings(findings))
  }
}
