#!/usr/bin/env bash

set -e

RUNTIME_FEATURE=$1
TMP_DIR=$(mktemp -d -t artemis-benchmark-XXX)

if [[ "$RUNTIME_FEATURE" == "with-snowbridge-runtime" ]]
then
    RUNTIME_DIR="runtime/snowbridge"
elif [[ "$RUNTIME_FEATURE" == "with-rococo-runtime" ]]
then
    RUNTIME_DIR="runtime/rococo"
else
    echo "Missing or invalid runtime feature argument. Pass either \"with-snowbridge-runtime\" or \"with-rococo-runtime\"."
    exit 1
fi

echo "Building runtime with features $RUNTIME_FEATURE,runtime-benchmarks"

FORCE_WASM_BUILD=$(date +%s) cargo build --release \
    --no-default-features \
    --features runtime-benchmarks,$RUNTIME_FEATURE

echo "Generating benchmark spec at $TMP_DIR/spec.json"

target/release/artemis build-spec > $TMP_DIR/spec.json
# Initialize dot-app account with enough DOT for benchmarks
DOT_MODULE_ENDOWMENT="[
    \"5EYCAe5jHEQsVPTRQqy6NCeG71Hz1EVXikZxTkr67fM8j2Rd\",
    1152921504606846976
]"
node ../test/scripts/helpers/overrideParachainSpec.js $TMP_DIR/spec.json \
    genesis.runtime.palletBalances.balances.0 "$DOT_MODULE_ENDOWMENT"

PALLETS="assets dot_app erc20_app eth_app frame_system pallet_balances pallet_timestamp verifier_lightclient"

echo "Generating weights module for $RUNTIME_DIR with pallets $PALLETS"

for pallet in $PALLETS
do
    target/release/artemis benchmark \
        --chain $TMP_DIR/spec.json \
        --execution wasm \
        --wasm-execution compiled \
        --pallet "${pallet}" \
        --extrinsic "*" \
        --repeat 20 \
        --steps 50 \
        --output $RUNTIME_DIR/src/weights/${pallet}_weights.rs
    echo "pub mod ${pallet}_weights;" >> $TMP_DIR/mod.rs
done

mv $TMP_DIR/mod.rs $RUNTIME_DIR/src/weights/mod.rs

echo "Done!"
