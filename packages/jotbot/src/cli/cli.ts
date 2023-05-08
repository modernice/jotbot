import { Command } from 'commander'
import { description, version } from '../../package.json'
import { withFindCmd } from './find'
import { withPosCmd } from './pos'
import { withMinifyCmd } from './minify'

/**
 * Creates a new Command Line Interface (CLI) instance with the specified
 * commands and options. The CLI is configured with a name, description, and
 * version from the package.json file. It includes the "find", "pos", and
 * "minify" commands using the {@link withFindCmd}, {@link withPosCmd}, and
 * {@link withMinifyCmd} functions respectively. Returns the configured CLI
 * instance.
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
