import { defineConfig } from 'tsup'

export default defineConfig({
  entryPoints: ['src/index.ts', 'src/commands/*'],
  external: ['fsevents'],
  format: ['cjs', 'esm'],
  outDir: 'dist',
  splitting: false,
  sourcemap: true,
  clean: true,
  dts: true,
})
