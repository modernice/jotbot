import ts from 'typescript'

/**
 * Creates a TypeScript {@link ts.SourceFile} object from the input string of
 * code, using the provided file name if available or a default name if not.
 */
export function parseCode(code: string, options?: { fileName?: string }) {
  return createSourceFile(options?.fileName ?? 'example.ts', code)
}

/**
 * Creates a TypeScript source file with the given filename and content. The
 * source file is created using the TypeScript Compiler API's {@link
 * ts.createSourceFile} function, with the latest script target and setting the
 * 'setParentNodes' parameter to true.
 */
export function createSourceFile(filename: string, content: string) {
  return ts.createSourceFile(filename, content, ts.ScriptTarget.Latest, true)
}

/**
 * Reads the content of a file given its path and returns it as a string. If the
 * file cannot be read, an error is thrown.
 */
export function readSource(file: string) {
  const content = ts.sys.readFile(file, 'utf8')
  if (!content) {
    throw new Error(`Could not read ${file}`)
  }
  return content
}

/**
 * Reads a file from the file system and returns a TypeScript source file object
 * representing the contents of the file. If the file cannot be read, an error
 * is thrown.
 */
export function readFile(file: string) {
  const content = readSource(file)
  return createSourceFile(file, content)
}
