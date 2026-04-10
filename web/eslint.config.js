import eslintPluginSvelte from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import tsParser from '@typescript-eslint/parser';
import sonarjs from 'eslint-plugin-sonarjs';
import unicorn from 'eslint-plugin-unicorn';
import prettierConfig from 'eslint-config-prettier';

export default [
	// Svelte files
	{
		files: ['**/*.svelte'],
		plugins: { svelte: eslintPluginSvelte, sonarjs, unicorn },
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: tsParser
			}
		},
		rules: {
			...eslintPluginSvelte.configs.recommended.rules,
			'sonarjs/cognitive-complexity': ['warn', 15],
			'no-duplicate-imports': 'error',
			'unicorn/prefer-string-replace-all': 'error',
			'unicorn/no-for-loop': 'error',
			'unicorn/no-typeof-undefined': 'error',
			'unicorn/prefer-at': 'error'
		}
	},
	// TS/JS files
	{
		files: ['**/*.{ts,js}'],
		plugins: { sonarjs, unicorn },
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				sourceType: 'module',
				ecmaVersion: 2020
			}
		},
		rules: {
			'sonarjs/cognitive-complexity': ['warn', 15],
			'no-duplicate-imports': 'error',
			'no-negated-condition': 'error',
			'max-params': ['error', 7],
			'default-param-last': 'error',
			'unicorn/prefer-string-replace-all': 'error',
			'unicorn/no-for-loop': 'error',
			'unicorn/prefer-global-this': 'error',
			'unicorn/no-typeof-undefined': 'error',
			'unicorn/prefer-string-raw': 'error',
			'unicorn/prefer-at': 'error'
		}
	},
	// Prettier disables conflicting rules (applied last)
	prettierConfig,
	// Ignore build output and generated files
	{
		ignores: ['.svelte-kit/', 'dist/', 'build/']
	}
];
