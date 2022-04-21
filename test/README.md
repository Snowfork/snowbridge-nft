# E2E tests

The E2E tests run against dockerized services.

## Requirements

1. Development environment for Rust and Substrate. See parachain [requirements](../parachain/README.md#requirements).
2. Development environment for Ethereum smart contracts.

   ```bash
   yarn global add truffle
   (cd ../ethereum && yarn install)
    ```

3. Docker and docker-compose

## Setup

Download dependencies:

```bash
yarn install
```

Start all services (parachain, relayer, ganache, etc):

```bash
scripts/start-services.sh
```

## Run Tests

```bash
yarn test
```

## Debugging

The `start-services.sh` script writes the following logs:

* parachain.log
* relayer.log
* ganache.log
