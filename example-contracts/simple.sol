//SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

contract SimpleStorage {
    int256 private storedData;

    function set(int256 x) public {
        storedData = x;
    }

    function get() public view returns (int256) {
        return storedData;
    }
}
