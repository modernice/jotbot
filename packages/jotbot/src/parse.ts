import ts from 'typescript'

/**
 * Parses a given string of code into a {@link ts.SourceFile} using the provided
 * file name, or defaults to 'example.ts' if no file name is specified. This
 * function utilizes TypeScript's own parser to create an abstract syntax tree
 * (AST) representation of the code for further analysis or manipulation.
 */
export function parseCode(code: string, options?: { fileName?: string }) {
  return createSourceFile(options?.fileName ?? 'example.ts', code)
}

/**
 * Generates a new TypeScript source file represented as a {@link ts.SourceFile}
 * object by parsing the provided content string using the latest script target.
 * The function assigns the given filename to the source file, which can be used
 * for error messages and in other parts of the TypeScript compiler API. If
 * syntax errors are present in the content, they will be captured within the
 * returned {@link ts.SourceFile}. The created source file also has its
 * 'setParentNodes' flag enabled, allowing for navigation of its node tree with
 * parent references.
 */
export function createSourceFile(filename: string, content: string) {
  return ts.createSourceFile(filename, content, ts.ScriptTarget.Latest, true)
}

/**
 * Reads the content of the specified file using UTF-8 encoding. If the file
 * cannot be read, it throws an error indicating the failure. The function
 * ensures that the file contents are returned as a string if successfully read.
 * If the file is not found or is unreadable, an exception is raised to prevent
 * further processing with invalid or missing data. This function is relevant
 * when source code needs to be loaded from a filesystem before further parsing
 * or manipulation.
 */
export function readSource(file: string) {
  const content = ts.sys.readFile(file, 'utf8')
  if (!content) {
    throw new Error(`Could not read ${file}`)
  }
  return content
}

/**
 * readFile reads the content of the specified file and creates a {@link
 * SourceFile} object representing it. It ensures that the file content is
 * properly read and handled, throwing an error if the file cannot be read. The
 * function leverages readFile from the underlying TypeScript system utilities
 * to handle file reading operations. The result is a {@link SourceFile} that
 * can be used in further TypeScript language processing or manipulation tasks.
 */
export function readFile(file: string) {
  const content = readSource(file)
  return createSourceFile(file, content)
}
