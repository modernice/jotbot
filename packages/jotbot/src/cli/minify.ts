import type { Command } from 'commander'
import type { TiktokenModel } from 'tiktoken'
import type { MinificationStep, MinifyFlags, MinifyToResult } from '../minify'
import { defaultMinificationSteps, isMinifyFlag, minifyTo } from '../minify'
import { readSource } from '../parse'
import type {
  WithFormatOption,
  WithSourceOption,
  WithVerboseOption,
} from './options'
import { formatOptions, parseFormat, verboseOption } from './options'
import { commaSeparated } from './utils'
import { createLogger } from './logger'
import { print } from './print'

interface Options
  extends WithSourceOption,
    WithFormatOption<'text' | 'json'>,
    WithVerboseOption {
  model: TiktokenModel
  tokens?: number
  steps?: (keyof MinifyFlags)[]
  printTokens?: boolean
}

/**
 * The `withMinifyCmd()` function adds a "minify" command to a given
 * Commander.js program. The minify command is used for minifying TypeScript or
 * JavaScript source code. It takes the source code as input, either directly or
 * from a specified file path, and applies a series of minification steps to
 * reduce the number of tokens in the output. The minification process can be
 * customized by specifying the maximum number of output tokens, the
 * minification steps to use, and other options.
 *
 * The function accepts a single parameter:
 *
 * - `program`: A Commander.js program instance to which the "minify" command
 * will be added.
 *
 * The "minify" command supports various options:
 *
 * - `code`: Optional TS/JS source code to minify.
 * - `--path`: Optional path to a TS/JS file to minify (instead of using
 * `code`).
 * - `--model`: OpenAI model used to determine the maximum output tokens
 * (default: 'text-davinci-003').
 * - `--tokens`: Maximum output tokens (overrides `--model`).
 * - `--steps`: Minification steps to use (comma-separated).
 * - `--print-tokens`: Return input and output tokens (requires
 * `--format="json"`).
 * - `--verbose`: Enable verbose logging.
 * - `--format`: Output format, either 'text' or 'json'.
 *
 * The function returns the modified Commander.js program instance.
 */
export function withMinifyCmd(program: Command) {
  const cmd = program
    .command('minify')
    .description('Minify TS/JS source code')
    .argument('[code]', 'TS/JS source code', '')
    .option('-p, --path <file>', 'Path to TS/JS file (instead of code)')
    .option(
      '-m, --model <model>',
      'OpenAI model to use',
      'gpt-3.5-turbo' as TiktokenModel,
    )
    .option(
      '-t, --tokens <number>',
      'Maximum output tokens (overrides `--model`)',
    )
    .option(
      '-s, --steps <steps>',
      'Minification steps (comma-separated)',
      commaSeparated({ validate: isMinifyFlag }),
      [] as (keyof MinifyFlags)[],
    )
    .option(
      '--print-tokens',
      'Return input and output tokens (requires `--format="json"`)',
      false,
    )
    .option(...verboseOption)

  formatOptions(['text', 'json'])(cmd)

  cmd.action(run)

  return program
}

function run(code: string, options: Options) {
  const { log, info } = createLogger(process.stderr, {
    enabled: options.verbose ?? false,
  })

  const model = options.model
  const maxTokens = options.tokens || maxTokensForModel(options.model)

  const stepsOption = options.steps?.length
    ? options.steps.reduce<Partial<MinifyFlags>>(
        (acc, flag) => ({ ...acc, [flag]: true }),
        {},
      )
    : undefined

  const steps = stepsOption ? [stepsOption] : defaultMinificationSteps

  if (options.path) {
    info(`File: ${options.path}`)
    code = readSource(options.path)
  }
  info(`Steps: ${steps.map((step) => Object.keys(step).join(',')).join(' > ')}`)

  if (options.path) {
    info(`Minifying ${options.path} ...`)
  } else {
    info(`Minifying code ...`)
  }

  const min = minifyTo(maxTokens, code, { model, steps })
  const { minified, inputTokens, tokens, steps: executedSteps } = min

  log('')
  info(`Max tokens: ${maxTokens}`)
  info(`Input: ${inputTokens.length} tokens`)
  info(`Minified: ${tokens.length} tokens`)
  if (executedSteps.length) {
    info(`Executed steps:`)
    for (const step of executedSteps) {
      log(`  ${formatStep(step)}`)
    }
  } else {
    info('Executed steps: none')
  }

  const format = parseFormat(options, 'text')

  log('')
  switch (format) {
    case 'text':
      print(minified)
      break
    case 'json':
      printJSON(min, { tokens: options.printTokens })
  }
}

const modelMaxTokens: { [m in TiktokenModel]?: number } & { default: 2049 } = {
  default: 2049,
  'gpt-4': 32768,
  'gpt-4-0314': 32768,
  'gpt-4-32k': 8192,
  'gpt-4-32k-0314': 8192,
  'gpt-3.5-turbo': 4096,
  'gpt-3.5-turbo-0301': 4096,
  'text-davinci-003': 4097,
  'text-davinci-002': 4097,
}

function maxTokensForModel(model: TiktokenModel) {
  return modelMaxTokens[model] || modelMaxTokens.default
}

function formatStep(step: MinificationStep) {
  return `> ${(Object.keys(step.flags) as (keyof MinifyFlags)[])
    .filter((flag) => step.flags[flag])
    .join(',')}`
}

function printJSON(
  result: MinifyToResult,
  options?: {
    tokens?: boolean
  },
) {
  const { minified, inputTokens, tokens, steps } = result
  const json = JSON.stringify(
    {
      steps: steps.map((step) => ({
        minified: step.minified,
        flags: step.flags,
        ...(options?.tokens
          ? {
              inputTokens: Array.from(step.inputTokens),
              tokens: Array.from(step.tokens),
            }
          : undefined),
      })),
      minified,
      ...(options?.tokens
        ? {
            inputTokens: Array.from(inputTokens),
            tokens: Array.from(tokens),
          }
        : undefined),
    },
    null,
    2,
  )
  print(json)
}
