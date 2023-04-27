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
