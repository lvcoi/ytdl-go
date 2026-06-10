module.exports = {
  ignorePatterns: [
    'internal/web/assets/app.js',
    'internal/web/assets/assets/**/*.js',
    'frontend/node_modules/',
  ],
  env: {
    browser: true,
    node: true,
    es2021: true
  },
  extends: ['eslint:recommended'],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true
    }
  },
  rules: {
    'no-undef': 'error',
    'no-unused-vars': 'warn',
    'no-empty': 'warn',
    'no-cond-assign': 'error',
    'no-prototype-builtins': 'warn',
    'no-constant-condition': 'warn'  // Add this line
  }
};
