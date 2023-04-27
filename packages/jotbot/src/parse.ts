import ts from 'typescript'

export function parseCode(code: string, options?: { fileName?: string }) {
  return createSourceFile(options?.fileName ?? 'example.ts', code)
}

export function createSourceFile(filename: string, content: string) {
  return ts.createSourceFile(filename, content, ts.ScriptTarget.Latest, true)
}

export function readSource(file: string) {
  const content = ts.sys.readFile(file, 'utf8')
  if (!content) {
    throw new Error(`Could not read ${file}`)
  }
  return content
}

export function readFile(file: string) {
  const content = readSource(file)
  return createSourceFile(file, content)
}
