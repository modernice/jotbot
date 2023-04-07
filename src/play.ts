import { useLogger } from '@modernice/openai'
import {
  playground,
  makeTemperatures,
  useTemperatures,
} from '@modernice/openai/play'
import { URL, fileURLToPath } from 'node:url'
import { generateForSymbol } from './generate.js'
import { patchFile } from './patch.js'
import { createRequire } from 'node:module'
import { exit } from 'node:process'
import { readFileSync } from 'node:fs'

const { resolve } = createRequire(import.meta.url)
const goCodePath = resolve('./fixtures/code.go')
const goCode = readFileSync(goCodePath, 'utf-8')

async function main() {
  await playground(async ({ logger: { log, info, success, error } }) => {
    // const temperatures = makeTemperatures(4, { min: 0, max: 0.4 })
    // const { run } = useTemperatures(temperatures)
    // const { logResults } = await run(async (temperature) => {
    //   return (
    //     await generateForSymbol('Of', goCode, {
    //       temperature,
    //       filepath: 'event/event.go',
    //       keywords: ['event-sourced', 'event-sourcing'],
    //     })
    //   ).content
    // })
    // logResults()

    const { content } = await generateForSymbol('Of', goCode, {
      filepath: 'event/event.go',
      keywords: ['event-sourcing', 'aggregate'],
      nouns: ['Aggregate', 'Event'],
    })

    const fixturePath = fileURLToPath(
      new URL('./fixtures/code.go', import.meta.url),
    )

    patchFile(fixturePath, 'Of', content)
  }, new URL('../.env', import.meta.url))
}

main().catch(useLogger().error)
