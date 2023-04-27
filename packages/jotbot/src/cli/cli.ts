import { Command } from 'commander'
import { description, version } from '../../package.json'
import { withFindCmd } from './find'
import { withPosCmd } from './pos'

export function createCLI() {
  const program = new Command('jotbot-es')
    .description(description)
    .version(version)

  withFindCmd(program)
  withPosCmd(program)

  return program
}
