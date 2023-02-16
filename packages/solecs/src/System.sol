// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import { ISystem } from "./interfaces/ISystem.sol";
import { IUint256Component } from "./interfaces/IUint256Component.sol";
import { IWorld } from "./interfaces/IWorld.sol";
import { SystemStorage } from "./SystemStorage.sol";

import { Ownable } from "./Ownable.sol";

// CSR
import { assignCSR } from "./turnstile-wrapper/ITurnstile.sol";

/**
 * System base contract
 */
abstract contract System is ISystem, Ownable {
  IUint256Component components;
  IWorld world;

  constructor(IWorld _world, address _components) {
    // @deprecated use SystemStorage.components() instead of components
    components = _components == address(0) ? _world.components() : IUint256Component(_components);
    // @deprecated use SystemStorage.world() instead of world
    world = _world;
    SystemStorage.init(world, components);
    if (address(_world) != address(0)) assignCSR(_world.getCSR_ID());
  }
}