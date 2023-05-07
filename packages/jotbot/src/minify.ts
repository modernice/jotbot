import { encoding_for_model } from '@dqbd/tiktoken'

/**
 * The defaultMinificationSteps variable is an array of objects that represent
 * the default order and configuration of minification steps. Each object
 * specifies which constructs (variables, classes, functions, etc.) should have
 * their associated comments removed during the minification process. This
 * variable is used as a fallback when no custom steps are provided to the
 * minifyTo function.
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
 * MinifyOptions is an interface that extends Partial<MinifyFlags> and adds an
 * optional computeTokens property. The computeTokens property, when set to
 * true, enables the computation of token counts for the input and minified code
 * in the resulting MinifyResult object.
 */
export interface MinifyOptions<ComputeTokens extends boolean = never>
  extends Partial<MinifyFlags> {
  /**
   * Determines whether to compute the token count of the input and minified
   * code when using {@link MinifyOptions}. If set to `true`, the resulting
   * {@link MinifyResult} will include `inputTokens` and `tokens` properties
   * representing the token counts.
   */
  computeTokens?: ComputeTokens
}

/**
 * MinifyFlags is an interface that represents the configuration object for
 * minification steps. It includes properties such as variables, functions,
 * classes, interfaces, properties, methods, and emptyLines that determine which
 * comments and empty lines to remove during the minification process.
 */
export interface MinifyFlags {
  /**
   * Indicates whether or not to remove comments associated with variables
   * during the minification process.
   * If set to true, comments preceding variable declarations will be removed
   * from the code.
   * This property is part of the {@link MinifyFlags} object and is used to
   * configure the minification steps.
   */
  variables: boolean
  /**
   * Represents the flag for minifying functions within {@link MinifyFlags}.
   * When set to true, it removes comments from function declarations in the
   * provided code during the minification process.
   */
  functions: boolean
  /**
   * The `classes` property of {@link MinifyFlags} determines whether class
   * comments should be removed or not during the minification process. If set
   * to `true`, class comments will be removed, otherwise they will be
   * preserved.
   */
  classes: boolean
  /**
   * Represents the flag for minifying interface declarations in the code. When
   * set to `true`, it removes comments associated with interface declarations
   * to reduce the size of the code. This property is used in conjunction with
   * other {@link MinifyFlags} properties to perform various minification steps
   * on the input code.
   */
  interfaces: boolean
  /**
   * Represents a flag that, when set to `true`, indicates the removal of
   * comments associated with properties within the code during minification.
   * This is done to reduce the token count and optimize the code further.
   * Used as part of the {@link MinifyFlags} configuration object.
   */
  properties: boolean
  /**
   * A flag that, when set to `true`, enables the removal of comments associated
   * with method declarations during the minification process. It is used within
   * the context of {@link MinifyFlags} to determine whether or not to perform
   * this specific optimization.
   */
  methods: boolean
  /**
   * Specifies whether empty lines should be removed from the code during
   * minification. When set to `true`, the {@link removeEmptyLines} function is
   * used to eliminate empty lines from the input code as part of the
   * minification process.
   */
  emptyLines: boolean
}

const _defaultFlags = minifyFlags()
/**
 * Checks if a string is a valid key of the {@link MinifyFlags} object,
 * indicating whether or not to perform a specific minification step on the
 * input code. Returns true if the string is a valid key, false otherwise.
 */
export function isMinifyFlag(s: string): s is keyof MinifyFlags {
  return s in _defaultFlags
}

/**
 * MinifyResult represents the result of a code minification operation, which
 * includes the minified code as a string and optionally the token counts of the
 * input and minified code. It is used as the return type of {@link minify}
 * function and {@link minifyTo} function.
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
 * MinificationStep represents a single step in the code minification process,
 * including the input code, minified code, and the specific minification flags
 * applied to remove comments and empty lines associated with variables,
 * functions, classes, interfaces, properties, and methods. It also includes the
 * token counts of the input and minified code if {@link
 * MinifyOptions.computeTokens} is set to true.
 */
export interface MinificationStep extends MinifyResult<true> {
  /**
   * Represents the input code for a {@link MinificationStep}. It is used to
   * store the original code before applying
   * minification flags in each step, allowing the {@link minifyTo} function to
   * keep track of the code state as it
   * iterates through the steps.
   */
  input: string
  /**
   * The `flags` property of {@link MinificationStep} represents the specific
   * minification actions
   * applied during the minification process. It includes removal of comments
   * and empty lines associated
   * with various code constructs like variables, functions, classes,
   * interfaces, properties, and methods.
   * This property helps in understanding the extent of minification performed
   * in each step and
   * assists in reaching the desired number of tokens.
   */
  flags: MinifyFlags
}

type Merge<A, B> = Omit<A, keyof B> & B

/**
 * MinifyToResult represents the return type of the minifyTo function, which
 * applies a series of minification steps to the input code and returns an
 * object containing the original and minified code, as well as token counts if
 * specified in the options.
 */
export type MinifyToResult = ReturnType<typeof minifyTo>

/**
 * Minifies the provided code according to the specified {@link MinifyFlags}
 * options, removing comments and empty lines associated with various code
 * constructs like variables, functions, classes, interfaces, properties, and
 * methods. If the resulting minified code has a token count exceeding the
 * specified maximum tokens, applies a series of minification steps until the
 * token count is below the maximum. Returns a {@link MinifyToResult} object
 * containing information about the minification process.
 */
export function minifyTo(
  maxTokens: number,
  code: string,
  options?: {
    steps?: Partial<MinifyFlags>[]
  },
) {
  const steps = options?.steps ?? defaultMinificationSteps
  const enc = encoding_for_model('text-davinci-003')
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
 * Minifies code by removing comments, empty lines, and other constructs based
 * on provided {@link MinifyFlags}. Can also compute token counts of input and
 * minified code if {@link MinifyOptions.computeTokens} is set to true.
 */
export function minify<ComputeTokens extends boolean = never>(
  code: string,
  options?: MinifyOptions<ComputeTokens>,
): MinifyResult<ComputeTokens> {
  const computeTokens = !!options?.computeTokens

  const enc = computeTokens ? encoding_for_model('text-davinci-003') : undefined
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
 * Removes empty lines from the provided code string and returns the modified
 * code. This function is used as part of the minification process when the
 * {@link MinifyFlags.emptyLines} flag is set to true.
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
