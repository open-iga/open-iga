import { defineConfig } from 'oxfmt';

export default defineConfig({
    ignorePatterns: ['**/package.json'],
    printWidth: 120,
    tabWidth: 4,
    useTabs: false,
    semi: true,
    singleQuote: true,
    trailingComma: 'all',
    bracketSpacing: true,
    bracketSameLine: false,
    arrowParens: 'always',
});
