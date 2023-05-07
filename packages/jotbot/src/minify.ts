import { encoding_for_model } from '@dqbd/tiktoken'

/**
 * The "defaultMinificationSteps" variable is an array of objects representing
 * the default order and levels of minification steps to be taken when minifying
 * code. Each object specifies which type of code elements should be removed in
 * that step, such as variables, properties, classes, interfaces, functions, and
 * methods. This variable is used as a fallback if no custom minification steps
 * are specified when calling the "minifyTo" function.
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
 * Describes the interface for MinifyOptions, which extends
 * Partial<MinifyFlags>. It allows for additional options to be passed to the
 * minification process, including the ability to compute tokens.
 */
export interface MinifyOptions<ComputeTokens extends boolean = never>
  extends Partial<MinifyFlags> {
  /**
   * The "MinifyOptions.computeTokens" property is an optional boolean value
   * that can be passed to the minification process. When set to true, it
   * enables the computation of input and output tokens for the minified code.
   */
  computeTokens?: ComputeTokens
}

/**
 * The "iface:MinifyFlags" interface defines a set of flags that control which
 * parts of a code string should be removed during minification, including
 * variables, functions, classes, interfaces, properties, methods, and empty
 * lines. These flags are used as options for the "minifyCode" function to
 * customize the minification process.
 */
export interface MinifyFlags {
  /**
   * The "MinifyFlags.variables" property is a boolean value that determines
   * whether or not to minify variables in the code during the minification
   * process. If set to `true`, any comments preceding variable declarations
   * will be removed from the code. If set to `false`, variable declarations
   * will not be minified.
   */
  variables: boolean
  /**
   * The `MinifyFlags.functions` property is a boolean flag that determines
   * whether to minify functions and methods in the code. When set to `true`,
   * the `minifyCode` function will remove all comments preceding function and
   * method declarations, effectively minifying them.
   */
  functions: boolean
  /**
   * The property "MinifyFlags.classes" is a boolean value that determines
   * whether to minify class declarations in the code during the minification
   * process. If set to true, all comments and white spaces related to class
   * declarations will be removed from the code.
   */
  classes: boolean
  /**
   * The `MinifyFlags.interfaces` property is a boolean flag that controls
   * whether or not interfaces should be removed during code minification. If
   * set to `true`, any exported or non-exported interfaces in the code will be
   * removed. If set to `false`, interfaces will be preserved.
   */
  interfaces: boolean
  /**
   * The "MinifyFlags.properties" property is an interface that defines a set of
   * boolean flags used to control which parts of a code string should be
   * removed during the minification process. These options include variables,
   * functions, classes, interfaces, properties, methods, and empty lines. These
   * flags are used as options for the "minifyCode" function to customize the
   * minification process.
   */
  properties: boolean
  /**
   * The `MinifyFlags.methods` property is a boolean value that determines
   * whether or not to minify method declarations in the code. If set to `true`,
   * any comments preceding a method declaration will be removed during the
   * minification process. If set to `false`, method declarations will not be
   * minified.
   */
  methods: boolean
  /**
   * Controls whether empty lines are removed during code minification. If set
   * to `true`, all empty lines in the code will be removed. If set to `false`,
   * empty lines will be preserved during minification. This option is enabled
   * by default.
   */
  emptyLines: boolean
}

const _defaultFlags = minifyFlags()
/**
 * Checks if a string is a valid key of the MinifyFlags object. Returns true if
 * the string is a valid key, false otherwise.
 */
export function isMinifyFlag(s: string): s is keyof MinifyFlags {
  return s in _defaultFlags
}

/**
 * Type "MinifyResult" represents the result of the minification process. It
 * contains the minified version of the input code and, if requested, the input
 * and output tokens. This type is used as the return type of the "minify"
 * function, which takes a string of code and returns a minified version of it
 * based on provided options, such as removing variables, functions, classes,
 * interfaces, properties, methods, and empty lines.
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
 * The "iface:MinificationStep" interface represents the result of a single step
 * in the minification process. It includes the input code, the minified code,
 * and the flags used during the step. Additionally, if requested, it includes
 * the input and output tokens of the step.
 */
export interface MinificationStep extends MinifyResult<true> {
  /**
   * The `MinificationStep.input` property is a string that represents the input
   * code to be minified. It is one of the properties of the `MinificationStep`
   * interface, which is a result object returned by the `minifyTo` function.
   * This property is used as an input to the `minify` function, which then
   * returns a minified version of the input code.
   */
  input: string
  /**
   * The `flags` property of the `MinificationStep` interface represents the
   * configuration options used during the minification process. It is an object
   * that contains boolean values for various minification options, such as
   * whether to minify variables, functions, classes, interfaces, properties,
   * and methods. It also includes an option to remove empty lines from the
   * code. These flags are used to control which parts of the code are modified
   * during the minification process.
   */
  flags: MinifyFlags
}

type Merge<A, B> = Omit<A, keyof B> & B

/**
 * Type "MinifyToResult" represents the return type of the function "minifyTo".
 * It contains the input code, minified code, input tokens, and output tokens
 * along with the steps executed during minification.
 */
export type MinifyToResult = ReturnType<typeof minifyTo>

/**
 * MinifyTo is a function that takes in a maximum number of tokens and a code
 * string, and returns the minified version of the code. It uses a set of
 * default minification steps which can be overridden by passing in custom steps
 * as an option. The function checks if the input code already has fewer tokens
 * than the maximum allowed and returns it unmodified if so. Otherwise, it
 * applies each step of minification until the token count is below the maximum
 * or there are no more steps left to apply. The function returns an object
 * containing both the input and minified versions of the code, as well as any
 * intermediate steps that were executed during the minification process.
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
 * The "minify()" function takes a string of code and returns a minified version
 * of it, with specified code elements removed based on the provided options.
 * The available options include removing variables, functions, classes,
 * interfaces, properties, methods, and empty lines. The function can also
 * compute input and output tokens if the "computeTokens" option is set to true.
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

/** Removes all empty lines from a given string of code. */
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
