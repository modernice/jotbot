import { Command } from 'commander'
import { description, version } from '../../package.json'
import { withFindCmd } from './find'

export function createCLI() {
  const program = new Command('jotbot-es')
    .description(description)
    .version(version)

  return withFindCmd(program)
}
