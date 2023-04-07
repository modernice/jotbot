import { getTokens, useLogger, writeMessage } from '@modernice/openai'

const DEFAULT_TEMPERATURE = 0
const DEFAULT_MAX_TOKENS = 512

export async function generateForSymbol(
  symbol: string,
  code: string,
  options?: {
    /**
     * Prepend the filepath as a comment to the code (to provide filesystem context).
     */
    filepath?: string

    /**
     * Instruct the model to consider the given keywords.
     */
    keywords?: string[]

    nouns?: string[]

    /**
     * @default 0
     */
    temperature?: number

    /**
     * @default 512
     */
    maxTokens?: number
  },
) {
  const instructions = `Write the documentation for the '${symbol}' type in GoDoc format, with symbols wrapped within brackets. Capitalize all proper nouns. Only output the documentation, not the input code. Do not include examples. Begin with "${symbol} is " or "${symbol} represents ".`

  const codeParts = [code]
  if (options?.filepath) {
    codeParts.unshift(`// ${options.filepath}\n`)
  }

  const history = [codeParts.join('\n')]

  if (options?.keywords?.length) {
    history.push(`Keywords: ${options.keywords.join(', ')}`)
  }

  const temperature = options?.temperature ?? DEFAULT_TEMPERATURE
  const maxTokens = computeMaxTokens(history.join('') + instructions, {
    max: options?.maxTokens,
  })

  let { answer: content } = await writeMessage(instructions, {
    temperature,
    maxTokens,
    history,
  })

  for (const noun of options?.nouns ?? []) {
    const re = new RegExp(`([^a-z0-9])${noun}([^a-z0-9-])`, 'gi')
    content = content.replaceAll(re, `$1${noun}$2`)
  }

  return {
    content,
  }
}

function computeMaxTokens(
  input: string,
  { total, ceil, max }: { total?: number; ceil?: number; max?: number },
) {
  total = total ?? 4096
  ceil = ceil ?? 100
  max = max ?? DEFAULT_MAX_TOKENS
  const tokens = getTokens(input).length
  const availableTokens = total - Math.ceil(tokens / ceil) * ceil
  return Math.min(availableTokens, max)
}

type UppercaseFirst<S extends string> = S extends `${infer First}${infer Rest}`
  ? `${Uppercase<First>}${Rest}`
  : S
