# polkadot-ethereum
Components for a Polkadot Ethereum Bridge

## Components

### Parachain
This folder includes our substrate parachain, as well as our bridge-specific pallets.

See [parachain/README.md](parachain/README.md)

### Relayer
This folder includes our Relayer daemon that will be run by relayers to watch and relay 2-way messages.

See [bridgerelayer/README.md](bridgerelayer/README.md)

### Prover
This folder includes our Prover daemon that will create proofs later used to verify cross-chain state.

See [prover/README.md](prover/README.md)
