# ERC721 Ethereum Contracts

This directory contains an archive of the ERC721 smart contract once utilized by the Polkadot-Ethereum Bridge.

## Development

Requirements:
* Node 14 LTS. See installation [instructions](https://www.digitalocean.com/community/tutorials/how-to-install-node-js-on-ubuntu-20-04#option-3-%E2%80%94-installing-node-using-the-node-version-manager).
* Yarn
* direnv: https://direnv.net/docs/installation.html

Install dependencies with yarn:

```console
$ yarn install
```

Create an `.envrc` file using [.envrc-example](.envrc-example) as a template. Note that deploying to ropsten network requires setting the INFURA_PROJECT_ID and ROPSTEN_PRIVATE_KEY environment variables.

Example:

```console
$ cp .envrc-example .envrc
$ direnv allow
```

## Testing

Run tests on the hardhat network:

```console
$ yarn test
```

## Deployment

### Local

Example: Run a local hardhat instance with deployments

```console
$ yarn hardhat node
```
