import ts from 'typescript'

const printer = ts.createPrinter({
  newLine: ts.NewLineKind.LineFeed,
  removeComments: false,
  omitTrailingSemicolon: true,
})

/**
 * Prints a given TypeScript {@link ts.Node} to a string using the TypeScript
 * compiler's printer utility. The function will throw an error if the node does
 * not have an associated source file. It returns the printed string
 * representation of the node without trailing semicolons and with comments
 * preserved.
 */
export function printNode(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file) {
    throw new Error('Node has no source file')
  }
  return printer.printNode(ts.EmitHint.Unspecified, node, file)
}

/**
 * printComment retrieves and formats the leading comments associated with a
 * given TypeScript {@link ts.Node}. If the node has synthetic leading comments,
 * it processes them using printSyntheticComments; otherwise, it defaults to
 * processing source file comments through printSourceFileComments. It returns a
 * formatted string representing the extracted comments. If the node lacks an
 * associated source file, an error is thrown.
 */
export function printComment(node: ts.Node) {
  const synthetic = ts.getSyntheticLeadingComments(node)
  if (synthetic != null) {
    return printSyntheticComments(synthetic)
  }
  return printSourceFileComments(node)
}

/**
 * Prints the text of synthetic leading comments associated with a given
 * TypeScript node or an array of synthesized comments. If provided a node, it
 * retrieves its synthetic leading comments; if provided an array, it uses the
 * comments directly. The resulting text is formatted as a block comment. This
 * function is particularly useful for emitting comments that were not
 * originally present in the source code but were added programmatically during
 * transformation processes. Returns a string representing the formatted comment
 * block. If no comments are found or provided, it returns an empty string
 * enclosed in comment delimiters.
 */
export function printSyntheticComments(
  nodeOrComments: ts.Node | ts.SynthesizedComment[],
) {
  const comments = Array.isArray(nodeOrComments)
    ? nodeOrComments
    : ts.getSyntheticLeadingComments(nodeOrComments) ?? []
  const body = comments.map((comment) => comment.text).join('\n')

  return `/*${body}*/`
}

/**
 * Prints the leading comments associated with a given {@link ts.Node} from its
 * source file. If the node has no associated source file or there are no
 * leading comments, the function will either throw an error or return an empty
 * string, respectively. The comments are retrieved directly from the source
 * text and concatenated into a single string, preserving their original format
 * and spacing.
 */
export function printSourceFileComments(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file) {
    throw new Error('Node has no source file')
  }
  const comments = ts.getLeadingCommentRanges(file.text, node.pos)
  if (!comments) {
    return ''
  }
  return comments
    .map((comment) => file.text.slice(comment.pos, comment.end))
    .join('\n')
}
