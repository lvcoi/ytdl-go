module.exports = {
  env: {
    browser: true,
    node: true,
    es2021: true
  },
  extends: ['eslint:recommended'],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
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
