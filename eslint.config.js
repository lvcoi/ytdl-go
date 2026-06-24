const {
    defineConfig,
    globalIgnores,
} = require("eslint/config");

const globals = require("globals");
const js = require("@eslint/js");

const {
    FlatCompat,
} = require("@eslint/eslintrc");

const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

module.exports = defineConfig([{
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

    extends: compat.extends("eslint:recommended"),

    rules: {
        "no-undef": "error",
        "no-unused-vars": "warn",
        "no-empty": "warn",
        "no-cond-assign": "error",
        "no-prototype-builtins": "warn",
        "no-constant-condition": "warn",
    },
}, globalIgnores([
    "internal/web/assets/app.js",
    "internal/web/assets/assets/**/*.js",
    "frontend/node_modules/",
]), globalIgnores(
    ["internal/web/assets/assets/", "**/node_modules/", "**/dist/", "**/build/"],
)]);
