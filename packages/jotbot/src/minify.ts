import type { TiktokenModel } from 'tiktoken'
import { encoding_for_model } from 'tiktoken'

/**
 * Defines an array of {@link MinifyFlags} objects representing the default
 * minification steps to be performed when minifying code.
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
 * MinifyOptions is an interface that provides optional configuration settings
 * for the minification process. It allows setting specific flags for removing
 * comments and empty lines from various code elements, as well as enabling the
 * computation of token counts for input and minified code using a specified
 * {@link TiktokenModel}.
 */
export interface MinifyOptions<ComputeTokens extends boolean = never>
  extends Partial<MinifyFlags> {
  /**
   * Determines whether the minification process should compute and return token
   * counts for the input and minified code. If `true`, {@link MinifyResult}
   * will include `inputTokens` and `tokens` properties with their respective
   * token counts. Set this option when you need to know the token counts after
   * minifying the code.
   */
  computeTokens?: ComputeTokens

  /**
   * Specifies the {@link TiktokenModel} to use when the `computeTokens` option
   * is set to `true`. This model is utilized to calculate the number of tokens
   * in the input and minified code. If `computeTokens` is not `true`, this
   * property should not be provided.
   */
  model?: [ComputeTokens] extends [true] ? TiktokenModel : never
}

/**
 * MinifyFlags represents an interface for specifying which code elements should
 * be minified, such as variables, functions, classes, interfaces, properties,
 * methods, and empty lines. This is used in the minification process to
 * configure the removal of comments and empty lines for a more compact output.
 */
export interface MinifyFlags {
  /**
   * Determines whether variable comments should be removed during the
   * minification process in {@link MinifyFlags}. Set to `true` to remove
   * variable comments, `false` otherwise.
   */
  variables: boolean
  /**
   * Indicates whether to remove comments associated with functions during
   * minification. When set to `true`, comments preceding function declarations
   * will be removed. This is used within the context of {@link MinifyFlags} to
   * determine which code elements should be minified.
   */
  functions: boolean
  /**
   * Indicates whether class-related comments should be removed during the
   * minification process. When set to `true`, it enables removing comments
   * associated with classes in the {@link MinifyFlags} configuration, resulting
   * in a more compact output.
   */
  classes: boolean
  /**
   * Indicates whether to remove comments associated with interface declarations
   * during the minification process. Set this property to `true` within a
   * {@link MinifyFlags} object to enable the removal of comments for
   * interfaces, improving the minification result by reducing the overall code
   * size.
   */
  interfaces: boolean
  /**
   * Specifies whether to remove comments associated with property declarations
   * in the code when minifying. Set to `true` by default in {@link
   * defaultMinificationSteps}. When used within {@link MinifyFlags}, it
   * controls the removal of such comments during the minification process.
   */
  properties: boolean
  /**
   * Specifies whether method comments should be removed during minification in
   * the {@link MinifyFlags}. When set to `true`, method comments will be
   * stripped from the code, reducing its size for more efficient processing.
   */
  methods: boolean
  /**
   * Determines whether empty lines should be removed from the code when
   * minifying. If set to `true`, all empty lines will be removed, resulting in
   * a more compact output. This is used within the context of {@link
   * MinifyFlags} to configure the minification process.
   */
  emptyLines: boolean
}

const _defaultFlags = minifyFlags()
/**
 * Determines whether the given string is a valid key of the {@link MinifyFlags}
 * interface.
 */
export function isMinifyFlag(s: string): s is keyof MinifyFlags {
  return s in _defaultFlags
}

/**
 * MinifyResult represents the result of a minification process, containing the
 * minified code and optionally the input tokens and output tokens if {@link
 * ComputeTokens} is set to true.
 */
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
 * MinificationStep represents a single step in the minification process,
 * containing the input code, minified code, input tokens, output tokens, and
 * the specific {@link MinifyFlags} applied during that step.
 */
export interface MinificationStep extends MinifyResult<true> {
  /**
   * Represents the input source code for a {@link MinificationStep}. The
   * `input` property serves as the starting point for minification in the
   * current step, and is updated as the minification process progresses through
   * each step.
   */
  input: string
  /**
   * Represents the set of minification options applied to the input code during
   * the current {@link MinificationStep}. It defines which aspects of the code
   * have been minified, such as variables, functions, classes, interfaces,
   * properties, methods, and empty lines.
   */
  flags: MinifyFlags
}

type Merge<A, B> = Omit<A, keyof B> & B

/**
 * MinifyToResult is the return type of the {@link minifyTo} function, which
 * takes a maximum token count, source code, and optional settings (including a
 * TiktokenModel and array of Partial<MinifyFlags>) as input. It iteratively
 * minifies the input code based on the provided steps, stopping when the token
 * count is below the maximum threshold. The returned object includes the input
 * code, minified code, input token count, final token count, and an array of
 * executed minification steps.
 */
export type MinifyToResult = ReturnType<typeof minifyTo>

/**
 * Minifies the input code to fit within the specified maximum number of tokens,
 * using the provided options and minification steps. Returns an object
 * containing the original input code, the minified code, their respective token
 * counts, and details of the executed minification steps.
 */
export function minifyTo(
  maxTokens: number,
  code: string,
  options?: {
    model?: TiktokenModel
    steps?: Partial<MinifyFlags>[]
  },
) {
  const model = options?.model ?? 'gpt-3.5-turbo'
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
 * Minifies the given code according to the provided options by removing
 * comments and empty lines. Returns the minified code and optionally the number
 * of input and output tokens when `computeTokens` option is enabled.
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
 * Removes empty lines from the given code string and trims any leading or
 * trailing whitespace.
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
