import type { Command } from 'commander'
import type { FinderOptions, SymbolType } from '..'
import { createFinder, printFindings, readSource } from '..'
import { isSymbol, symbolTypes } from '../symbols'
import { createLogger } from './logger'
import { commaSeparated } from './utils'
import type {
  WithFormatOption,
  WithSourceOption,
  WithVerboseOption,
} from './options'
import { formatOptions, verboseOption } from './options'
import { out } from './print'

const { log: print } = out

interface Options
  extends FinderOptions,
    WithFormatOption<'json' | 'list'>,
    WithSourceOption,
    WithVerboseOption {}

export function withFindCmd(program: Command) {
  const cmd = program
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
    .option(...verboseOption)
    .addHelpText(
      'after',
      '\nDefault --symbols:\n  ["func", "var", "class", "method", "iface", "prop"]',
    )

  formatOptions(['list', 'json'])(cmd)

  cmd.action(run)

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
