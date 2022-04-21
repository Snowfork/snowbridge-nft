// SPDX-License-Identifier: MIT
pragma solidity >=0.6.2;

import "@openzeppelin/contracts/math/SafeMath.sol";
import "./Decoder.sol";

contract EthereumApp {
    using SafeMath for uint256;
    using Decoder for bytes;

    uint256 public nonce;
    uint256 public totalETH;

    enum AppEventTags { SendETH, SendERC20 }

    event AppEvent(uint _tag, bytes _data);
    event Unlock(bytes _sender, address _recipient, uint256 _amount);

    constructor() public {
        nonce = 0;
    }

    function sendETH(
        bytes32 _recipient
    )
        public
        payable
    {
        require(msg.value > 0, "Value of transaction must be positive");

        // Increment locked Ethereum counter by this amount
        totalETH = totalETH.add(msg.value);
        // Increment global nonce
        nonce = nonce.add(1);

        bytes memory data = encodeSendData(msg.sender, _recipient, address(0), msg.value, nonce);
        emit AppEvent(uint(AppEventTags.SendETH), data);
    }

    function encodeSendData(
        address _sender,
        bytes32 _recipient,
        address _tokenAddr,
        uint256 _amount,
        uint256 _nonce
    )
        internal
        pure
        returns(bytes memory)
    {
        return abi.encode(_sender, _recipient, _tokenAddr, _amount, _nonce);
    }


    function submit(bytes memory _data, bytes memory _signature)
        public
    {
        require(_data.length == 84, "Data must contain 84 bytes for a successful decoding");

        // Decode sender bytes
        bytes memory sender = _data.slice(0, 32);
        // Decode recipient address
        address payable recipient = _data.sliceAddress(32);
        // Deocde amount int256
        bytes memory amountBytes = _data.slice(32 + 20, 32);
        uint256 amount = amountBytes.decodeUint256();

        sendETH(recipient, amount);
        emit Unlock(sender, recipient, amount);
    }

    function sendETH(
        address payable _recipient,
        uint256 _amount
    )
        internal
    {
        require(_amount > 0, "Must unlock a positive amount");
        require(totalETH > _amount, "ETH token balances insufficient to fulfill the unlock request");

        totalETH = totalETH.sub(_amount);
        _recipient.transfer(_amount);
    }
}
