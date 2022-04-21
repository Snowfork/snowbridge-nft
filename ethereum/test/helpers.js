const rlp = require("rlp");
const { keccakFromHexString, keccak } = require("ethereumjs-util");
const MerkleTree = require("merkletreejs").MerkleTree;

const MerkleProof = artifacts.require("MerkleProof");
const Bitfield = artifacts.require("Bitfield");
const ScaleCodec = artifacts.require("ScaleCodec");
const ValidatorRegistry = artifacts.require("ValidatorRegistry");
const MMRVerification = artifacts.require("MMRVerification");
const Blake2b = artifacts.require("Blake2b");
const BeefyLightClient = artifacts.require("BeefyLightClient");

const fixture = require('./fixtures/beefy-fixture-data.json');

let lazyLinked = false;
const lazyLinkLibraries = async _ => {
  if (lazyLinked) {
    return;
  }
  const merkleProof = await MerkleProof.new();
  const bitfield = await Bitfield.new();
  const scaleCodec = await ScaleCodec.new();

  await ValidatorRegistry.link(merkleProof); // 860624903cbc2e721b1f7f70307ce6b5fe
  await BeefyLightClient.link(bitfield); // ce679fb3689ba2b0521c393162ea0c3c96$
  await BeefyLightClient.link(scaleCodec); // 7cdc5241ea8d29c91205423c213999ecf3
  lazyLinked = true;
}

const initValidatorRegistry = async (validatorRoot, numOfValidators, validatorSetID) => {
  await lazyLinkLibraries()

  validatorRegistry = await ValidatorRegistry.new(
    validatorRoot,
    numOfValidators,
    validatorSetID
  );

  return validatorRegistry;
}

const makeBasicCommitment = (messages) => {
  let encoded = ethers.utils.defaultAbiCoder.encode(
    ['tuple(address target, uint64 nonce, bytes payload)[]'],
    [messages]
  )
  return ethers.utils.solidityKeccak256(["bytes"], [encoded])
}

const makeIncentivizedCommitment = (messages) => {
  let encoded = ethers.utils.defaultAbiCoder.encode(
    ['tuple(address target, uint64 nonce, uint256 fee, bytes payload)[]'],
    [messages]
  )
  return ethers.utils.solidityKeccak256(["bytes"], [encoded])
}

const deployAppWithMockChannels = async (deployer, channels, appContract, ...appContractArgs) => {
  const app = await appContract.new(
    ...appContractArgs,
    {
      inbound: channels[0],
      outbound: channels[1],
    },
    {
      inbound: channels[0],
      outbound: channels[1],
    },
    {
      from: deployer,
    }
  );

  return app;
}

const deployBeefyLightClient = async _ => {
  this.validatorsMerkleTree = createMerkleTree(fixture.validatorPublicKeys);
  const root = this.validatorsMerkleTree.getHexRoot()

  const validatorRegistry = await initValidatorRegistry(root,
    fixture.validatorPublicKeys.length, fixture.startingValidatorSetID);
  const mmrVerification = await MMRVerification.new();
  const blake2b = await Blake2b.new();

  const beefyLightClient = await BeefyLightClient.new(
    validatorRegistry.address,
    mmrVerification.address,
    blake2b.address,
    0
  );

  await validatorRegistry.transferOwnership(beefyLightClient.address)

  return beefyLightClient;
}

function signatureSubstrateToEthereum(sig) {
  const recoveryId0 = web3.utils.hexToNumber(`0x${sig.slice(130)}`);
  const newRecoveryId0 = web3.utils.numberToHex(recoveryId0 + 27);
  const res = sig.slice(0, 130).concat(newRecoveryId0.slice(2));

  return res;
}

function createMerkleTree(leavesHex) {
  const leavesHashed = leavesHex.map(leaf => keccakFromHexString(leaf));
  const merkleTree = new MerkleTree(leavesHashed, keccak, { sort: true });

  return merkleTree;
}

async function mine(n) {
  for (let i = 0; i < n; i++) {
    web3.currentProvider.send({
      jsonrpc: '2.0',
      method: 'evm_mine',
      params: [],
      id: new Date().getTime()
    }, (err, res) => { });
  }
}

const addressBytes = (address) => Buffer.from(address.replace(/^0x/, ""), "hex");

const encodeLog = (log) => {
  return rlp.encode([log.address, log.topics, log.data]).toString("hex")
}

const ChannelId = {
  Basic: 0,
  Incentivized: 1,
}

const hexPrefix = /^(0x)/i

const mergeKeccak256 = (left, right) =>
  '0x' + keccakFromHexString('0x' + left.replace(hexPrefix, "") + right.replace(hexPrefix, ''), 256).toString('hex')

const PREFIX = "VM Exception while processing transaction: ";

async function tryCatch(promise, type, message) {
  try {
    await promise;
    throw null;
  }
  catch (error) {
    assert(error, "Expected an error but did not get one");
    if (message) {
      assert(error.message === (PREFIX + type + ' ' + message), "Expected error '" + PREFIX + type + ' ' + message +
        "' but got '" + error.message + "' instead");
    } else {
      assert(error.message.startsWith(PREFIX + type), "Expected an error starting with '" + PREFIX + type +
        "' but got '" + error.message + "' instead");
    }
  }
};

module.exports = {
  deployAppWithMockChannels,
  deployBeefyLightClient,
  createMerkleTree,
  signatureSubstrateToEthereum,
  mine,
  addressBytes,
  ChannelId,
  encodeLog,
  mergeKeccak256,
  catchRevert: async (promise, message) => await tryCatch(promise, "revert", message),
  catchOutOfGas: async (promise, message) => await tryCatch(promise, "out of gas", message),
  catchInvalidJump: async (promise, message) => await tryCatch(promise, "invalid JUMP", message),
  catchInvalidOpcode: async (promise, message) => await tryCatch(promise, "invalid opcode", message),
  catchStackOverflow: async (promise, message) => await tryCatch(promise, "stack overflow", message),
  catchStackUnderflow: async (promise, message) => await tryCatch(promise, "stack underflow", message),
  catchStaticStateChange: async (promise, message) => await tryCatch(promise, "static state change", message),
};
