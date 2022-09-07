# Cadence Import
Imports a contract and that contracts dependencies recursively. Also updates your `flow.json`. The contracts import statements will also be updated to local relative paths.
This should produce a local set of contracts in the format expected by the [VSCode Cadence Extension](https://github.com/onflow/vscode-cadence) and [Flow Emulator](https://github.com/onflow/flow-emulator).

## Installation
Run `go install github.com/jrkhan/cadence-import`

Make sure you have an up to date install of the [flow-cli](https://github.com/onflow/flow-cli)

## Usage
Navigate to a empty directory for your new project.
Run `flow init` to create your `flow.json`.

Add the name and address of at least one contract you'd like to import to `flow.json`, you'll need to supply an alias for mainnet or testnet.

Run `cadence-import get {ContractName}` to get contracts src from the version currently deployed on chain.

You may also supply the network and address as flags instead of adding them to `flow.json`.

`cadence-import -a 547f177b243b4d80 -n testnet TopShotMarketV3`

