import { Command } from 'commander'
import { description, version } from '../../package.json'
import { withFindCmd } from './find'
import { withPosCmd } from './pos'
import { withMinifyCmd } from './minify'

export function createCLI() {
  const program = new Command('jotbot-es')
    .description(description)
    .version(version)

  withFindCmd(program)
  withPosCmd(program)
  withMinifyCmd(program)

  return program
}
