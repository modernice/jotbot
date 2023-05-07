import { fileURLToPath } from 'node:url'
import { createCLI } from './cli'

/**
 * The main function initializes and runs the command line interface (CLI) by
 * parsing the provided arguments.
 * If no arguments are provided, it uses the default process arguments.
 * 
 * @param args - An optional array of strings representing command line
 * arguments.
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
