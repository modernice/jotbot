import { Command } from 'commander'
import { description, version } from '../../package.json'
import { withFindCmd } from './find'
import { withPosCmd } from './pos'
import { withMinifyCmd } from './minify'

/**
 * Initializes and configures a Command Line Interface (CLI) for the jotbot-ts
 * application, incorporating various subcommands such as find, pos, and minify.
 * Returns the configured {@link Command} instance ready for execution.
 */
export function createCLI() {
  const program = new Command('jotbot-ts')
    .description(description)
    .version(version)

  withFindCmd(program)
  withPosCmd(program)
  withMinifyCmd(program)

  return program
}
