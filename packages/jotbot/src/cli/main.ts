import { fileURLToPath } from 'node:url'
import { createCLI } from './cli'

/**
 * Initializes the command-line interface with the given arguments and starts
 * the parsing process, utilizing the {@link createCLI} utility. If invoked
 * directly, it will automatically use arguments from the current process.
 */
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
