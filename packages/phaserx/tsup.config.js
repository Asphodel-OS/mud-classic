import { defineConfig } from 'tsup'

export default defineConfig({
  clean: true,
  dts: true,
  entryPoints: ['src/index.ts'],
  format: ['cjs', 'esm'],
  splitting: false,
  sourcemap: true,
  target: 'es2020',
})
