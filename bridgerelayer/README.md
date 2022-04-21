# Bridgerelayer

Relayer service that streams transactions from blockchain networks, packages data into messages, and sends the packets to the correlated bridge component.

Note: the bridgerelayer is currently in a boilerplate/architectural design state, it's not functional yet.

## Setup

```bash
export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off

make install
```

For testing, start a local Ethereum network and deploy the Bank contract by following the set up instructions [here](../ethereum/README.md).

A sample configuration file is provided at test_config.json. Update the operator and private key fields with valid values from your local Ethereum network.

## Usage

```bash
# Check that the binary was successfully installed
bridgerelayer -h

# Start the relayer
bridgerelayer init
```

You should see a message similar to
```bash
INFO[0000] Connected to Ethereum chain ID 5777          
INFO[0000] Subscribed to app 0xC4cE93a5699c68241fc2fB503Fb0f21724A624BB 
```

You can send a `sendEth` transaction to the Bank contract with default values via the sendEth script located in polkadot-ethereum/ethereum/scripts/sendEth.js

```bash
# Send the transaction
truffle exec sendEth.js

# You should see the transaction in the bridgerelayer
INFO[0007] Witnessed tx 0x22c26a2d423bcc9622daba9410f5bdee1d047ec2e8be5c112a01b64224dbea5e on app 0xC4cE93a5699c68241fc2fB503Fb0f21724A624BB 
```

Currently, the relayer logs the packet instead of sending it directly to the bridge. It should look similar to
```bash
INFO[0007] Send packet:
{[196 206 147 165 105 156 104 36 31 194 251 80 63 176 242 23 36 166 36 187 0 0 0 0 0 0 0 0 0 0 0 0] {{[249 1 250 148 196 206 147 165 105 156 104 36 31 194 251 80 63 176 242 23 36 166 36 187 225 160 38 100 19 190 87 0 206 141 213 172 107 154 125 251 171 233 155 62 69 202 233 166 138 194 117 120 88 113 11 64 26 56 185 1 192 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 96 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 192 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 48 116 97 114 103 101 116 32 97 112 112 108 105 99 97 116 105 111 110 39 115 32 117 110 105 113 117 101 32 115 117 98 115 116 114 97 116 101 32 105 100 101 110 116 105 102 105 101 114 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 7 115 101 110 100 69 84 72 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 151 17 95 110 32 4 215 180 204 214 185 213 171 52 227 9 9 224 246 18 205 49 70 82 77 77 56 80 69 105 87 88 89 97 120 55 114 112 83 54 88 52 88 90 88 49 97 65 65 120 83 87 120 49 67 114 75 84 121 114 86 89 104 86 50 52 102 103 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 10 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 72 0 0 0 0 0 0 0 0 0] {[240 226 133 71 131 90 191 98 228 147 243 241 207 55 102 89 130 95 112 225 127 230 50 235 143 143 114 76 15 7 173 94] [157 1 12 70 186 234 126 129 134 227 42 230 20 207 194 178 194 58 35 113 16 85 195 47 164 221 242 239 100 159 75 44 35 195 162 146 204 63 203 91 149 186 154 126 132 92 126 7 63 253 109 238 50 16 94 3 109 21 52 29 85 202 202 78 0]}}}}
```


## Previous work

Thanks to Chainsafe for their work on [ChainBridge](https://github.com/ChainSafe/ChainBridge), a event-based bridge relayer that this project is based on.

