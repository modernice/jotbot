import ts from 'typescript'

const printer = ts.createPrinter({
  newLine: ts.NewLineKind.LineFeed,
  removeComments: false,
  omitTrailingSemicolon: true,
})

export function printFile(file: ts.SourceFile) {
  return printer.printFile(file)
}

export function printNode(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file)
    throw new Error('Node has no source file')
  return printer.printNode(ts.EmitHint.Unspecified, node, file)
}

export function printComment(node: ts.Node) {
  const synthetic = ts.getSyntheticLeadingComments(node)
  if (synthetic != null)
    return printSyntheticComments(synthetic)

  return printSourceFileComments(node)
}

export function printSyntheticComments(
  nodeOrComments: ts.Node | ts.SynthesizedComment[],
) {
  const comments = Array.isArray(nodeOrComments)
    ? nodeOrComments
    : ts.getSyntheticLeadingComments(nodeOrComments) ?? []
  const body = comments.map(comment => comment.text).join('\n')

  return `/*${body}*/`
}

export function printSourceFileComments(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file)
    throw new Error('Node has no source file')
  const comments = ts.getLeadingCommentRanges(file.text, node.pos)
  if (comments == null)
    return ''
  return comments
    .map(comment => file.text.slice(comment.pos, comment.end))
    .join('\n')
}
