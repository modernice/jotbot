import { encoding_for_model } from '@dqbd/tiktoken'

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

export interface MinifyOptions<ComputeTokens extends boolean = never>
  extends Partial<MinifyFlags> {
  computeTokens?: ComputeTokens
}

/**
 * MinifyFlags is an interface that defines the flags used to control the
 * minification process in the code. It specifies whether to minify variables,
 * functions, classes, interfaces, properties, methods, and empty lines. Each
 * flag is a boolean value that determines whether to minify or remove the
 * corresponding code element during the minification process.
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

export interface MinificationStep extends MinifyResult<true> {
  input: string
  flags: MinifyFlags
}

type Merge<A, B> = Omit<A, keyof B> & B

export type MinifyToResult = ReturnType<typeof minifyTo>

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
