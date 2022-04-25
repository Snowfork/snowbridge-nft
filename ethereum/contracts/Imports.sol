// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.5;

import "@snowfork/snowbridge-contracts/contracts/utils/MerkleProof.sol";
import "@snowfork/snowbridge-contracts/contracts/utils/Bitfield.sol";
import "@snowfork/snowbridge-contracts/contracts/ParachainLightClient.sol";
import "@snowfork/snowbridge-contracts/contracts/BasicInboundChannel.sol";
import "@snowfork/snowbridge-contracts/contracts/IncentivizedInboundChannel.sol";
import "@snowfork/snowbridge-contracts/contracts/BasicOutboundChannel.sol";
import "@snowfork/snowbridge-contracts/contracts/IncentivizedOutboundChannel.sol";
import "@snowfork/snowbridge-contracts/contracts/ETHApp.sol" as ETHApp;
import "@snowfork/snowbridge-contracts/contracts/DOTApp.sol" as DOTApp;