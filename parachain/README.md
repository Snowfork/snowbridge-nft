# Snowbridge ERC721 Pallets <!-- omit in toc -->

A Polkadot parachain for bridging arbitrary data from and to Ethereum.

- [Development](#development)
  - [Requirements](#requirements)
  - [Build](#build)
  - [Test](#test)
- [Configuration](#configuration)

## Development

Follow these steps to prepare your local environment for Substrate development.

### Requirements

Refer to the instructions at the
[Substrate Developer Hub](https://substrate.dev/docs/en/knowledgebase/getting-started/#manual-installation).

To add context to the above instructions, the parachain is known to compile with the following versions of Rust:

- stable: 1.58
- nightly: 1.60.0-nightly

### Build

To build:
```bash
cargo build --workspace --release
```

### Test

To test:
```bash
cargo test --workspace --features runtime-benchmarks
```