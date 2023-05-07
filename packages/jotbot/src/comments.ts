import ts from 'typescript'
import type { RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'
import { findNode } from './nodes'
import { printComment } from './print'

/**
 * The `Comments` type represents a collection of comments indexed by their
 * associated identifiers. It provides various utility functions for handling
 * comments in TypeScript source files, such as parsing identifiers, formatting
 * comments, and updating comments associated with specific nodes. The
 * `Comments` type allows for easy manipulation and management of comments
 * within a TypeScript project.
 */
export type Comments = Record<RawIdentifier, string>

/**
 * The `getIdentifiers()` function takes a `Comments` object as input and
 * returns an array of parsed identifiers. It first retrieves the raw
 * identifiers from the input object, then maps each raw identifier to its
 * parsed form using the `parseIdentifier()` function.
 */
export function getIdentifiers(comments: Comments) {
  return getRawIdentifiers(comments).map(parseIdentifier)
}

/**
 * The `getRawIdentifiers()` function takes a `Comments` object as input and
 * returns an array of raw identifiers, which are the keys of the input object.
 */
export function getRawIdentifiers(comments: Comments) {
  return Object.keys(comments) as Array<keyof Comments>
}

/**
 * The `patchComments()` function takes a TypeScript source file and a record of
 * comments, where the keys are identifiers and the values are the associated
 * comments. It updates the source file by inserting or modifying comments for
 * each identifier found in the record. If an identifier in the record is not
 * found in the source file, an error is thrown.
 */
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

/**
 * The `formatComment()` function takes a comment string and optional options
 * object as input, and returns a formatted comment string. The options object
 * can include two properties: `maxLen`, which sets the maximum line length
 * (default is 80), and `enclose`, which determines if the comment should be
 * enclosed in `/* *\/` (default is false). The function splits the input
 * comment into lines based on the maximum line length, and formats each line
 * with a leading ` *`. If `enclose` is true, the formatted comment will be
 * enclosed in `/* *\/`.
 */
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

/**
 * The `getNodeComments()` function retrieves the leading comments of a given
 * TypeScript node. It returns an object containing an array of comments and the
 * printed text of the node.
 */
export function getNodeComments(node: ts.Node) {
  const comments = ts.getSyntheticLeadingComments(node) ?? []
  const text = printComment(node)

  return {
    comments,
    text,
  }
}

/**
 * The `getInsertPosition()` function returns the line and character position of
 * a given {@link ts.Node} in its source file. This is useful for determining
 * where to insert or update comments within the code.
 */
export function getInsertPosition(node: ts.Node) {
  return ts.getLineAndCharacterOfPosition(node.getSourceFile(), node.getStart())
}

/**
 * The `updateNodeComments()` function updates the leading comments of a given
 * TypeScript node with the specified array of comments. It combines the
 * existing synthetic leading comments with the new comments, preserving their
 * order and formatting.
 */
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
