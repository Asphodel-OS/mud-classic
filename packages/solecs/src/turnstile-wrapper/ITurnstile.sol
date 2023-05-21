// SPDX-License-Identifier: Unlicense
pragma solidity >=0.8.0;

// Turnstile contract address fixed on mainnet
address constant TurnstileAddress = 0xEcf044C5B4b867CFda001101c617eCd347095B44;
uint256 constant CantoMainChainID = 7700;
uint256 constant CantoTestChainID = 7701;

function isCanto() view returns (bool) {
  return (block.chainid == CantoMainChainID) || (block.chainid == CantoTestChainID);
}

// registers a new CSR to world, returns nft ID
function registerCSR() returns (uint256) {
  if (isCanto()) {
    return ITurnstile(TurnstileAddress).register(tx.origin);
  }

  return 0;
}

// assigns new components to world's turnsile
function assignCSR(uint256 id) {
  if (isCanto()) {
    ITurnstile(TurnstileAddress).assign(id);
  }
}

interface ITurnstile {
  function register(address) external returns (uint256);

  function assign(uint256) external returns(uint256);
}
