### Cadence Import
Imports contracts and that contracts recursive dependencies locally and updates your `flow.json`. Imports in the contract source will also be updated to a local relative path. 
This might be useful for bootstrapping projects using the VSCode Cadence extension and/or Flow emulator.

### Usage
To install, run `go install github.com/jrkhan/cadence-import`
Navigate to a new directory.
Create a new folder and run `flow init` to create your `flow.json`.

Create or import a `registry.json` (format is identical to flow.json) with at least one contract you'd like to import, you'll need to supply an alias for mainnet or testnet.

Run `cadence-import get {ContractName}` to get contracts src from the version currently deployed on chain.

