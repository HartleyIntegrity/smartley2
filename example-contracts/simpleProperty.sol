//SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

contract PropertyRental {
    struct Property {
        address owner;
        uint256 price;
        bool isRented;
    }

    mapping(uint256 => Property) public properties;
    uint256 public propertyCount;

    mapping(uint256 => address) public tenants;

    event PropertyAdded(uint256 indexed propertyId, uint256 price, address indexed owner);
    event PropertyRented(uint256 indexed propertyId, address indexed tenant);
    event RentalCompleted(uint256 indexed propertyId);

    function addProperty(uint256 price) public {
        properties[propertyCount] = Property(msg.sender, price, false);
        emit PropertyAdded(propertyCount, price, msg.sender);
        propertyCount++;
    }

    function rentProperty(uint256 propertyId) public payable {
        require(propertyId < propertyCount, "Invalid property ID");
        require(!properties[propertyId].isRented, "Property is already rented");
        require(msg.value == properties[propertyId].price, "Incorrect rental price");

        properties[propertyId].owner.transfer(msg.value);
        properties[propertyId].isRented = true;
        tenants[propertyId] = msg.sender;

        emit PropertyRented(propertyId, msg.sender);
    }

    function completeRental(uint256 propertyId) public {
        require(msg.sender == properties[propertyId].owner || msg.sender == tenants[propertyId], "Not authorized");
        properties[propertyId].isRented = false;
        tenants[propertyId] = address(0);

        emit RentalCompleted(propertyId);
    }
}