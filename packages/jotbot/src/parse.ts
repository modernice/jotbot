import ts from 'typescript'
import { printFile } from './print'

export function parseCode(
  code: string,
  options?: {
    fileName?: string
  },
) {
  return createSourceFile(options?.fileName ?? 'example.ts', code)
}

export function createSourceFile(filename: string, content: string) {
  return ts.createSourceFile(
    filename,
    content,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TS,
  )
}

export function readFile(fileName: string) {
  const content = ts.sys.readFile(fileName)
  if (!content)
    throw new Error(`Could not read ${fileName}`)

  return createSourceFile(fileName, content)
}

export function cloneFile(file: ts.SourceFile) {
  const content = printFile(file)
  return createSourceFile(file.fileName, content)
}
