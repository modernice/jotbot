import type { Command } from 'commander'
import {
  findNode,
  getInsertPosition,
  isRawIdentifier,
  parseCode,
  readSource,
} from '..'
import { createLogger } from './logger'
import type { WithSourceOption, WithVerboseOption } from './options'
import { verboseOption } from './options'
import { print } from './print'

interface Options extends WithSourceOption, WithVerboseOption {}

/**
 * Registers the `pos` command within a given {@link Command} instance, which
 * extracts and prints the position of a specified node from TypeScript or
 * JavaScript source code. Accepts an identifier to locate the desired node and
 * an optional source code string or file path. If the identifier is invalid or
 * the node cannot be found, the process exits with an error message. The
 * function enhances the provided {@link Command} instance with this new
 * capability and returns it.
 */
export function withPosCmd(program: Command) {
  program
    .command('pos')
    .description('Extract node position from TS/JS source code')
    .argument('<identifier>', 'Identifier of the node (e.g. "func:foo")')
    .argument('[code]', 'TS/JS source code', '')
    .option('-p, --path <file>', 'Path to TS/JS file (instead of code)')
    .option(...verboseOption)
    .action(run)

  return program
}

function run(identifier: string, code: string, options: Options) {
  const { log, info } = createLogger(process.stderr, {
    enabled: options.verbose ?? false,
  })

  if (options.path) {
    info(`File: ${options.path}`)
    code = readSource(options.path)
  }

  if (!isRawIdentifier(identifier)) {
    log(`Invalid identifier: ${identifier}`)
    process.exit(1)
  }

  const file = parseCode(code)

  const node = findNode(file, identifier)
  if (!node) {
    log(`Node not found: ${identifier}`)
    process.exit(1)
  }

  const pos = getInsertPosition(node.commentTarget)

  log('')
  print(JSON.stringify(pos))
}
