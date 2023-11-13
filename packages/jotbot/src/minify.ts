import type { TiktokenModel } from 'tiktoken'
import { encoding_for_model } from 'tiktoken'

/**
 * Specifies a sequence of minification configurations to be applied
 * successively during the code minification process. Each configuration is a
 * partial set of {@link MinifyFlags}, indicating which aspects of the code are
 * subject to minification at that particular step.
 */
export const defaultMinificationSteps = [
  {
    variables: true,
    properties: true,
  },
  {
    classes: true,
    interfaces: true,
  },
  {
    functions: true,
    methods: true,
  },
] as Partial<MinifyFlags>[]

const variableCommentsRE =
  /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:export\s+)?(?:var|let|const)\s+)/g
const functionCommentsRE =
  /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:export\s+)?(?:function)\s+)/g
const classCommentsRE =
  /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:export\s+)?(?:class)\s+)/g
const interfaceCommentsRE =
  /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:export\s+)?(?:interface)\s+)/g
const propertyCommentsRE =
  /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:\w+\s*:\s*\w+))/g
const methodCommentsRE = /(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)\s*(?=(?:\w+\s*\())/g
const emptyLineRE = /^\s*[\r\n]+/gm

/**
 * Defines configuration options for the minification process. `MinifyOptions`
 * is an extension of {@link MinifyFlags} and includes optional properties that
 * determine the behavior during minification. If `computeTokens` is set, it
 * indicates whether token counts should be computed, requiring a specified
 * `model` when `true`. The `model` property is only relevant when
 * `computeTokens` is `true`, in which case it must be a valid {@link
 * TiktokenModel}.
 */
export interface MinifyOptions<ComputeTokens extends boolean = never>
  extends Partial<MinifyFlags> {
  computeTokens?: ComputeTokens

  model?: [ComputeTokens] extends [true] ? TiktokenModel : never
}

/**
 * Represents a set of boolean flags used to configure the minification process.
 * Each flag indicates whether a specific type of code element—variables,
 * functions, classes, interfaces, properties, methods, or empty lines—should
 * be included in the minification process. When a flag is set to `true`, the
 * corresponding code element will be considered for minification.
 */
export interface MinifyFlags {
  variables: boolean
  functions: boolean
  classes: boolean
  interfaces: boolean
  properties: boolean
  methods: boolean
  emptyLines: boolean
}

const _defaultFlags = minifyFlags()
/**
 * Determines whether the provided string argument is a valid key of the {@link
 * MinifyFlags} interface. Returns `true` if the argument corresponds to a
 * property name within {@link MinifyFlags}, otherwise `false`. This check is
 * case-sensitive and does not account for optional properties.
 */
export function isMinifyFlag(s: string): s is keyof MinifyFlags {
  return s in _defaultFlags
}

export type MinifyResult<
  ComputeTokens extends boolean = never,
  WithTokens = true extends ComputeTokens ? true : false,
> = Merge<
  { minified: string },
  true extends WithTokens
    ? { inputTokens: Uint32Array; tokens: Uint32Array }
    : {}
>

/**
 * Represents an individual step in the minification process, containing the
 * state of the code and associated tokens before and after the step is applied.
 * Each `MinificationStep` includes the source code prior to minification
 * (`input`), the resulting code after applying minification (`minified`), and a
 * set of flags (`flags`) indicating which minification options were used.
 * Additionally, it provides the original and final tokens as `Uint32Array`
 * instances when token computation is enabled, offering insight into how the
 * code's token representation changes through this step. This type is integral
 * to understanding and debugging each stage of the minification pipeline,
 * ensuring that developers can track transformations and optimize their code
 * effectively.
 */
export interface MinificationStep extends MinifyResult<true> {
  input: string
  flags: MinifyFlags
}

type Merge<A, B> = Omit<A, keyof B> & B

export type MinifyToResult = ReturnType<typeof minifyTo>

/**
 * Performs a sequence of code minification steps on the given source code until
 * the token count is less than or equal to the specified maximum number of
 * tokens. The minification process includes optional transformations such as
 * removal of comments, whitespace, and unused code based on the provided
 * configuration flags. If a {@link TiktokenModel} is passed in the options, it
 * is used to encode the source and minified code into tokens for comparison.
 * The function returns a {@link MinifyToResult}, which includes the original
 * and minified code strings, their respective token arrays if token computation
 * was requested, and an array of executed {@link MinificationStep}s detailing
 * each transformation applied during the minification process.
 */
export function minifyTo(
  maxTokens: number,
  code: string,
  options?: {
    model?: TiktokenModel
    steps?: Partial<MinifyFlags>[]
  },
) {
  let model = options?.model ?? 'gpt-3.5-turbo'
  if ((model as any) === 'gpt-4-1106-preview') {
    model = 'gpt-4'
  }

  const steps = options?.steps ?? defaultMinificationSteps
  const enc = encoding_for_model(model)
  const executedSteps = [] as MinificationStep[]

  let inputTokens = enc.encode(code)
  let tokens = inputTokens

  if (inputTokens.length <= maxTokens) {
    return { input: code, minified: code, inputTokens, tokens, steps: [] }
  }

  let input = code
  let minified = code
  for (const step of steps) {
    const flags = minifyFlags(step)
    const min = minify(input, {
      ...flags,
      model,
      computeTokens: true,
    })

    executedSteps.push({
      ...min,
      input,
      flags,
    })

    input = min.minified
    minified = min.minified
    inputTokens = min.inputTokens
    tokens = min.tokens

    if (min.tokens.length <= maxTokens) {
      break
    }
  }

  return { input, minified, inputTokens, tokens, steps: executedSteps }
}

/**
 * Transforms a string of code into a more compact version by removing or
 * altering elements based on specified {@link MinifyFlags}. If the
 * `computeTokens` option is set, the function also computes and returns the
 * tokenized representation of both the input and the minified code using the
 * provided {@link TiktokenModel}. The result is returned as a {@link
 * MinifyResult}, which includes the minified string and, optionally, the token
 * arrays.
 */
export function minify<ComputeTokens extends boolean = never>(
  code: string,
  options?: MinifyOptions<ComputeTokens>,
): MinifyResult<ComputeTokens> {
  const computeTokens = !!options?.computeTokens
  const model = options?.model ?? 'gpt-3.5-turbo'

  const enc = computeTokens ? encoding_for_model(model) : undefined
  const inputTokens = computeTokens ? enc!.encode(code) : undefined

  const minified = minifyCode(code, options)
  const tokens = computeTokens ? enc!.encode(minified) : undefined

  return {
    minified,
    ...(computeTokens ? { inputTokens, tokens } : undefined),
  } as any
}

function minifyCode(code: string, options?: Partial<MinifyFlags>) {
  if (options?.variables ?? true) {
    code = code.replace(variableCommentsRE, '')
  }

  if (options?.functions ?? true) {
    code = code.replace(functionCommentsRE, '')
  }

  if (options?.classes ?? true) {
    code = code.replace(classCommentsRE, '')
  }

  if (options?.interfaces ?? true) {
    code = code.replace(interfaceCommentsRE, '')
  }

  if (options?.properties ?? true) {
    code = code.replace(propertyCommentsRE, '')
  }

  if (options?.methods ?? true) {
    code = code.replace(methodCommentsRE, '')
  }

  if (options?.emptyLines ?? true) {
    code = removeEmptyLines(code)
  }

  return code
}

/**
 * Removes all empty lines from the provided code string, returning the cleaned
 * text with no leading or trailing whitespace.
 */
export function removeEmptyLines(code: string) {
  return code.replace(emptyLineRE, '').trim()
}

function minifyFlags(flags?: Partial<MinifyFlags>): MinifyFlags {
  return {
    variables: flags?.variables ?? false,
    functions: flags?.functions ?? false,
    classes: flags?.classes ?? false,
    interfaces: flags?.interfaces ?? false,
    properties: flags?.properties ?? false,
    methods: flags?.methods ?? false,
    emptyLines: flags?.emptyLines ?? false,
  }
}
