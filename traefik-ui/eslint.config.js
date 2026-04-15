import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'src/components/ui/**'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      'no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: 'react',
              importNames: ['useEffect'],
              message:
                'Direct useEffect is banned. Use derived state, event handlers, React Query, `key`-prop remounts, or useMountEffect() for genuine mount-only side effects.',
            },
          ],
        },
      ],
      'no-restricted-syntax': [
        'error',
        {
          selector: "CallExpression[callee.name='useEffect']",
          message:
            'Direct useEffect is banned. See src/hooks/useMountEffect.ts or refactor to derived state / event handlers / React Query.',
        },
        {
          selector: "CallExpression[callee.object.name='React'][callee.property.name='useEffect']",
          message: 'React.useEffect is banned. Use useMountEffect() if genuinely mount-only.',
        },
      ],
    },
  },
)
