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
import { out } from './print'

const { log: print } = out

interface Options extends WithSourceOption, WithVerboseOption {}

/**
 * The `withPosCmd()` function is used to create and configure a `pos` command
 * for a given Commander.js program instance. The `pos` command extracts the
 * position of a specified node identifier from TypeScript or JavaScript source
 * code. The function accepts a single parameter:
 * 
 * - `program`: A Commander.js program instance.
 * 
 * The `pos` command takes two arguments:
 * 
 * - `<identifier>`: The identifier of the node to find (e.g. "func:foo").
 * - `[code]`: The TypeScript or JavaScript source code (optional).
 * 
 * Additionally, the `pos` command provides two options:
 * 
 * - `-p, --path <file>`: Path to a TypeScript or JavaScript file (instead of
 * providing the source code as an argument).
 * - `-v, --verbose`: Enable verbose output.
 * 
 * Upon successful execution, the `pos` command prints the extracted position of
 * the specified node in JSON format.
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
