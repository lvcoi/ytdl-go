const globals = require("globals");
const js = require("@eslint/js");

module.exports = [
    js.configs.recommended,
    {
        languageOptions: {
            globals: {
                ...globals.browser,
                ...globals.node,
            },
            ecmaVersion: "latest",
            sourceType: "module",
            parserOptions: {
                ecmaFeatures: {
                    jsx: true,
                },
            },
        },
        rules: {
            "no-undef": "error",
            "no-unused-vars": "warn",
            "no-empty": "warn",
            "no-cond-assign": "error",
            "no-prototype-builtins": "warn",
            "no-constant-condition": "warn",
        },
        ignores: [
            "internal/web/assets/app.js",
            "internal/web/assets/assets/**/*.js",
            "frontend/node_modules/",
            "internal/web/assets/assets/",
            "**/node_modules/",
            "**/dist/",
            "**/build/",
            "**/*.md"
        ]
    }
];
