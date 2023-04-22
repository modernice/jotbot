import { defineConfig } from 'vite'
import dts from 'vite-plugin-dts'
import { readPackageJSON } from 'pkg-types'

export default defineConfig(async () => {
  const pkg = await readPackageJSON()

  function isExternal(dep: string) {
    return (
      dep.startsWith('node:') ||
      [
        ...Object.keys(pkg.dependencies ?? {}),
        ...Object.keys(pkg.peerDependencies ?? {}),
      ].includes(dep)
    )
  }

  return {
    plugins: [dts({ insertTypesEntry: true })],

    build: {
      lib: {
        entry: './src/index.ts',
        name: 'JotBot',
        fileName: (format) => `index.${format === 'es' ? 'mjs' : format}`,
        formats: ['es', 'cjs'],
      },

      rollupOptions: {
        external: isExternal,
      },
    },
  }
})
