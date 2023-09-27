// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import { SystemConfig } from "src/L1/SystemConfig.sol";
import { SuperchainConfig } from "src/L1/SuperchainConfig.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { ResourceMetering } from "src/L1/ResourceMetering.sol";
import { Constants } from "src/libraries/Constants.sol";

contract SystemConfig_GasLimitLowerBound_Invariant is Test {
    SystemConfig public config;

    function setUp() external {
        Proxy proxy = new Proxy(msg.sender);
        SystemConfig configImpl = new SystemConfig();

        vm.prank(msg.sender);
        proxy.upgradeToAndCall(
            address(configImpl),
            abi.encodeCall(
                configImpl.initialize,
                (
                    address(0xbeef), // owner
                    address(0), // _superchainConfig
                    SystemConfig.GasConfig({
                        overhead: 2100, // overhead
                        scalar: 1000000 // scalar
                     }),
                    bytes32(hex"abcd"), // batcher hash
                    30_000_000, // gas limit
                    address(1), // unsafe block signer
                    Constants.DEFAULT_RESOURCE_CONFIG(), // resource config
                    0, //_startBlock
                    address(0), // _batchInbox
                    address(0), // _challenger
                    SystemConfig.Addresses({ // _addrs
                        l1CrossDomainMessenger: address(0),
                        l1ERC721Bridge: address(0),
                        l1StandardBridge: address(0),
                        l2OutputOracle: address(0),
                        optimismPortal: address(0),
                        optimismMintableERC20Factory: address(0)
                    })
                )
            )
        );

        config = SystemConfig(address(proxy));

        // Set the target contract to the `config`
        targetContract(address(config));
        // Set the target sender to the `config`'s owner (0xbeef)
        targetSender(address(0xbeef));
        // Set the target selector for `setGasLimit`
        // `setGasLimit` is the only function we care about, as it is the only function
        // that can modify the gas limit within the SystemConfig.
        bytes4[] memory selectors = new bytes4[](1);
        selectors[0] = config.setGasLimit.selector;
        FuzzSelector memory selector = FuzzSelector({ addr: address(config), selectors: selectors });
        targetSelector(selector);
    }

    /// @custom:invariant The gas limit of the `SystemConfig` contract can never be lower
    ///                   than the hard-coded lower bound.
    function invariant_gasLimitLowerBound() external {
        assertTrue(config.gasLimit() >= config.minimumGasLimit());
    }
}
