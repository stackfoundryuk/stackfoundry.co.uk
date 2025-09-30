 
import { dirname } from "path";
import { fileURLToPath } from "url";

import { FlatCompat } from "@eslint/eslintrc";
import unusedImportsPlugin from "eslint-plugin-unused-imports";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),

  {
    ignores: [
      "node_modules/**",
      ".next/**",
      "out/**",
      "build/**",
      "next-env.d.ts",
    ],
  },

  {
    plugins: {
      "unused-imports": unusedImportsPlugin,
    },
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: __dirname,
      },
    },
    rules: {
      // General formatting / style
      quotes: ["error", "double", { avoidEscape: true }],
      semi: ["error", "always"],
      "comma-dangle": ["error", "always-multiline"],
      "object-curly-spacing": ["error", "always"],
      "array-bracket-spacing": ["error", "never"],
      "keyword-spacing": ["error", { before: true, after: true }],
      "space-before-blocks": ["error", "always"],

      // Imports
      "no-duplicate-imports": "error",
      "sort-imports": [
        "error",
        {
          ignoreCase: true,
          ignoreDeclarationSort: true, // handled by import/order
          ignoreMemberSort: false,
          memberSyntaxSortOrder: ["none", "all", "multiple", "single"],
        },
      ],
      "import/order": [
        "error",
        {
          groups: [
            "builtin",
            "external",
            "internal",
            ["parent", "sibling", "index"],
          ],
          "newlines-between": "always",
          alphabetize: { order: "asc", caseInsensitive: true },
        },
      ],

      // Unused code cleanup
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "no-unused-vars": "off",
      "unused-imports/no-unused-imports": "error",
      "unused-imports/no-unused-vars": [
        "warn",
        { vars: "all", varsIgnorePattern: "^_", argsIgnorePattern: "^_" },
      ],

      // Best practices
      eqeqeq: ["error", "always"],
      "no-console": ["warn", { allow: ["warn", "error"] }],

      // TypeScript naming conventions
      "@typescript-eslint/naming-convention": [
        "error",

        // Variables (const, let, var)
        {
          selector: "variable",
          format: ["camelCase"],
          leadingUnderscore: "allow",
        },

        // Functions (including exported functions)
        {
          selector: "function",
          format: ["PascalCase"],
        },

        // Types, Interfaces, Classes, Enums
        {
          selector: "typeLike",
          format: ["PascalCase"],
        },

        // Enum members
        {
          selector: "enumMember",
          format: ["PascalCase"],
        },

        // Components (functions starting with uppercase)
        {
          selector: "variable",
          types: ["function"],
          format: ["PascalCase"],
        },
      ],
    },
  },
];

export default eslintConfig;
