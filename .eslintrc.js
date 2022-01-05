module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      modules: true,
    },
    ecmaVersion: 2021,
    sourceType: 'module',
  },
  extends: [
    'plugin:@typescript-eslint/recommended',
    'prettier',
    'plugin:prettier/recommended',
  ],
  rules: {
    'import/extensions': [
      'error',
      'ignorePackages',
      {
        js: 'never',
        // jsx: 'never',
        ts: 'never',
        // tsx: 'never',
      },
    ],
    'no-use-before-define': 'off',
    'no-shadow': 'off',
    'max-len': ['error', { code: 120 }],
    'semi': ['error', 'never'],
    'quotes': ['error', 'single'],
    // 'jsx-quotes': ['error', 'prefer-double'],
    'quote-props': ['error', 'consistent'],
    'comma-dangle': [
      'error',
      {
        arrays: 'always-multiline',
        objects: 'always-multiline',
        imports: 'always-multiline',
        exports: 'always-multiline',
        functions: 'ignore',
      },
    ],
    'object-curly-newline': [
      'error',
      {
        consistent: true,
      },
    ],
    'no-confusing-arrow': ['error', { allowParens: true }],
    'arrow-parens': ['error', 'always'],
    'no-param-reassign': ['error', { props: false }],
    'no-var': 'error',
    'no-eval': 'error',
    'no-unused-vars': 'off',
    'eqeqeq': ['error', 'always', { null: 'ignore' }],
    // TODO: or with max-len limit
    'arrow-body-style': 'off',
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      },
    ],
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-explicit-any': ['error', { ignoreRestArgs: true }],
    '@typescript-eslint/no-non-null-assertion': 'error',
    '@typescript-eslint/no-use-before-define': 'warn',
    '@typescript-eslint/no-shadow': ['error'],
    '@typescript-eslint/no-redeclare': ['error'],
  },
  settings: {
    'import/resolver': {
      typescript: {},
      node: {
        paths: ['src'],
        extensions: ['.js', '.json', '.ts'],
      },
    },
  },
}
