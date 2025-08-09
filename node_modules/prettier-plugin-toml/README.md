# prettier-plugin-toml ![npm bundle size](https://img.shields.io/bundlephobia/min/prettier-plugin-toml) ![npm bundle size](https://img.shields.io/bundlephobia/minzip/prettier-plugin-toml)

> An opinionated `toml` formatter plugin for [Prettier][]

Prettier is an opinionated code formatter. It enforces a consistent style by parsing your code and re-printing, taking various rules into account.

This plugin adds support for `toml` through [taplo][].

## Notice

This plugin is still under development, its printer just wraps [taplo][]'s default printer.
Of course it should just work, but may not match [prettier][]'s format sometimes.

## Requirements

`prettier-plugin-toml` is an evergreen module. 🌲 This module requires an [LTS](https://github.com/nodejs/Release) Node version (v16.0.0+).

## Install

Using npm:

```sh
# npm
npm i -D prettier prettier-plugin-toml

# yarn
yarn add -D prettier prettier-plugin-toml
```

## Usage

Once installed, [Prettier plugins](https://prettier.io/docs/en/plugins.html) must be added to `.prettierrc`:

```json
{
  "plugins": ["prettier-plugin-toml"]
}
```

Then:

```sh
# npx
npx prettier --write foo.toml

# yarn
yarn prettier --write foo.toml
```

## Parser Options

```ts
interface PrettierOptions {
  // 'Align consecutive entries vertically.'
  alignEntries: boolean // default `false`
  // Align consecutive comments after entries and items vertically. This applies
  // to comments that are after entries or array items.
  alignComments: boolean // default `true`
  // Expand arrays to multiple lines that exceed the maximum column width.
  arrayAutoExpand: boolean // default `true`
  // Collapse arrays that don't exceed the maximum column width and don't
  // contain comments.
  arrayAutoCollapse: boolean // default `true`
  // Omit white space padding from single-line arrays.
  compactArrays: boolean // default `true`
  // Omit white space padding from the start and end of inline tables.
  compactInlineTables: boolean // default `false`
  // Omit white space around `=`.
  compactEntries: boolean // default `false`
  // Indent based on tables and arrays of tables and their subtables, subtables
  // out of order are not indented.
  indentTables: boolean // default `false`
  // Indent entries under tables.
  indentEntries: boolean // default `false`
  // Alphabetically reorder keys that are not separated by empty lines.
  reorderKeys: boolean // default `false`
  // The maximum number of allowed blank lines between entries and tables.
  allowedBlankLines: number // integer, default `1`
}
```

## Sponsors and Backers

[![Sponsors and Backers](https://raw.githubusercontent.com/1stG/static/master/sponsors.svg)](https://github.com/sponsors/JounQin)

### Sponsors

| 1stG                                                                                                                   | RxTS                                                                                                                   | UnTS                                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| [![1stG Open Collective sponsors](https://opencollective.com/1stG/organizations.svg)](https://opencollective.com/1stG) | [![RxTS Open Collective sponsors](https://opencollective.com/rxts/organizations.svg)](https://opencollective.com/rxts) | [![UnTS Open Collective sponsors](https://opencollective.com/unts/organizations.svg)](https://opencollective.com/unts) |

### Backers

| 1stG                                                                                                                | RxTS                                                                                                                | UnTS                                                                                                                |
| ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| [![1stG Open Collective backers](https://opencollective.com/1stG/individuals.svg)](https://opencollective.com/1stG) | [![RxTS Open Collective backers](https://opencollective.com/rxts/individuals.svg)](https://opencollective.com/rxts) | [![UnTS Open Collective backers](https://opencollective.com/unts/individuals.svg)](https://opencollective.com/unts) |

## Changelog

Detailed changes for each release are documented in [CHANGELOG.md](./CHANGELOG.md).

## License

[MIT][] © [Ray][]@[mk1.io][]

[mk1.io]: https://mk1.io
[mit]: http://opensource.org/licenses/MIT
[prettier]: https://prettier.io
[ray]: https://github.com/so1ve
[taplo]: https://github.com/tamasfe/taplo
