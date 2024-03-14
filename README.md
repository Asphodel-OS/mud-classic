# MUD

<div align="center">
<img src="public/logo512.png" width="200" style="margin: 0 0 30px 0;" alt="MUD logo" />
<p>MUD Classic - Simplified Engine for Autonomous Worlds</p>
</div>

<p align="center">
  <a aria-label="license MIT" href="https://opensource.org/licenses/MIT">
    <img alt="" src="https://img.shields.io/badge/License-MIT-yellow.svg">
  </a>
</p>

MUD Classic, is simplified fork of MUDv1, focusing on the use case of complex onchain dapps with large (rather than realtime) latency overheads.

MUD is a framework for complex Ethereum applications. It standardizes the way data is stored on-chain. It adds some conventions for organizing data and logic and abstracts away low-level complexities so you can focus on the features of your app.

MUD Classic is MIT-licensed, open source and free to use if you can figure it out.. (it's missing parts)

## Packages

MUD consists of several core libraries. They can be used independently, but are best used together.

| Package                                                                                                    | Version                                                                                                                                 |
| ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| **[@mud-classic/cli](/packages/cli)** <br />Command line interface for types, testing, deployment and more | [![npm version](https://img.shields.io/npm/v/@mud-classic/cli.svg)](https://www.npmjs.org/package/@mud-classic/cli)                     |
| **[@mud-classic/recs](/packages/recs)** <br />TypeScript Reactive Entity Component System library          | [![npm version](https://img.shields.io/npm/v/@mud-classic/recs.svg)](https://www.npmjs.org/package/@mud-classic/recs)                   |
| **[@mud-classic/solecs](/packages/solecs)** <br />Solidity Entity Component System library                 | [![npm version](https://img.shields.io/npm/v/@mud-classic/solecs.svg)](https://www.npmjs.org/package/@mud-classic/solecs)               |
| **[@mud-classic/std-contracts](/packages/std-contracts)** <br />Solidity standard library                  | [![npm version](https://img.shields.io/npm/v/@mud-classic/std-contracts.svg)](https://www.npmjs.org/package/@mud-classic/std-contracts) |
| **[@mud-classic/utils](/packages/utils)** <br />Arbitrary utilities you may find useful..                  | [![npm version](https://img.shields.io/npm/v/@mud-classic/utils.svg)](https://www.npmjs.org/package/@mud-classic/utils)                 |

### Local development setup

1. Install go (required to build [packages/services](packages/services/)): [https://go.dev/doc/install](https://go.dev/doc/install)

2. Install the foundry toolkit (required to build and test MUD solidity packages): [https://getfoundry.sh/](https://getfoundry.sh/)

3. Clone the MUD monorepo

```bash
git clone https://github.com/Asphodel-OS/mud-classic
```

4. Install MUD dependencies and setup local environment

```bash
cd mud-classic && pnpm i
```

### Pull requests

MUD follows the [conventional commit specification](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages and PR titles. Please keep the scope of your PR small (rather open multiple small PRs than one huge PR) and follow the conventional commit spec.

## License

MUD is open-source software [licensed as MIT](LICENSE).
