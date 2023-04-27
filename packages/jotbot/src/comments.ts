import ts from 'typescript'
import type { RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'
import { findNode } from './nodes'
import { printComment } from './print'

export type Comments = Record<RawIdentifier, string>

export function getIdentifiers(comments: Comments) {
  return getRawIdentifiers(comments).map(parseIdentifier)
}

export function getRawIdentifiers(comments: Comments) {
  return Object.keys(comments) as Array<keyof Comments>
}

export function patchComments(
  file: ts.SourceFile,
  comments: Record<string, string>,
) {
  const fileName = file.fileName
  const identifiers = getRawIdentifiers(comments)

  for (const identifier of identifiers) {
    const comment = formatComment(comments[identifier])

    if (!updateComment(file, identifier, comment)) {
      throw new Error(`Could not find identifier ${identifier} in ${fileName}`)
    }
  }
}

function updateComment(
  tree: ts.Node,
  identifier: RawIdentifier,
  comment: string,
) {
  const node = findNode(tree, parseIdentifier(identifier))
  if (!node) {
    return false
  }

  updateNodeComments(node.commentTarget, [comment])
  return true
}

export function formatComment(
  comment: string,
  options?: {
    /**
     * @default 80
     */
    maxLen?: number

    /**
     * @default false
     */
    enclose?: boolean
  },
): string {
  const lines = splitIntoLines(comment, (options?.maxLen ?? 80) - 3)
  let formatted = `*\n${lines.map((line) => ` * ${line}`).join('\n')}\n `

  if (options?.enclose) {
    formatted = `/*${formatted}*/`
  }

  return formatted
}

export function getNodeComments(node: ts.Node) {
  const comments = ts.getSyntheticLeadingComments(node) ?? []
  const text = printComment(node)

  return {
    comments,
    text,
  }
}

export function updateNodeComments(node: ts.Node, comments: string[]) {
  ts.setSyntheticLeadingComments(node, [
    ...(ts.getSyntheticLeadingComments(node) ?? []),
    ...comments.map(
      (comment, i) =>
        ({
          kind: ts.SyntaxKind.MultiLineCommentTrivia,
          text: comment.trim() + (i === comments.length - 1 ? '\n ' : ''),
          hasTrailingNewLine: true,
          pos: -1,
          end: -1,
        } satisfies ts.SynthesizedComment),
    ),
  ])
}

function splitIntoLines(str: string, maxLen: number): string[] {
  const out: string[] = []

  const paras: string[] = str.split('\n\n')
  for (let i = 0; i < paras.length; i++) {
    const para: string = paras[i]
    const lines: string[] = splitParagraph(para, maxLen)
    out.push(...lines)
    if (i < paras.length - 1) {
      out.push('')
    }
  }

  return out
}

function splitParagraph(str: string, maxLen: number): string[] {
  const words: string[] = str.split(' ')

  const lines: string[] = []
  let line = ''
  for (const word of words) {
    if (line.length + word.length > maxLen) {
      lines.push(line)
      line = ''
    }
    line += `${word} `
  }
  lines.push(line.trim())

  return lines
}
