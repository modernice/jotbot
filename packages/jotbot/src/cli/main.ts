import { createCLI } from './cli'
import { fileURLToPath } from 'node:url'

export function main(args: readonly string[] = process.argv.slice(2)) {
  const cli = createCLI()
  cli.parse(args, { from: 'user' })
}

if (import.meta.url.startsWith('file:')) {
  const modulePath = fileURLToPath(import.meta.url)
  if (process.argv[1] === modulePath) {
    main()
  }
}
