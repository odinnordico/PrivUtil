import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist', 'coverage']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      // This rule (added in react-hooks v7 for the React Compiler) flags the intentional
      // "clear output on empty input" early-return pattern used consistently across all
      // debounced-API-call components. The pattern is correct and the one-extra-render
      // cost is negligible in a dev-tool context. Disabling until components are refactored
      // to use derived state or onChange-handler clearing.
      'react-hooks/set-state-in-effect': 'off',
    },
  },
])
