// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./TenancyAgreement.sol";

contract RentPayment {
    TenancyAgreement private tenancyAgreement;
    mapping (address => mapping (uint256 => uint256)) private rentPayments;

    constructor(address _tenancyAgreementAddress) {
        tenancyAgreement = TenancyAgreement(_tenancyAgreementAddress);
    }

    event RentPaid(address indexed tenant, uint256 agreementId, uint256 amount);

    function makeRentPayment(uint256 _agreementId) public payable {
        (, address tenant, uint256 rentAmount, , , ) = tenancyAgreement.getAgreement(_agreementId);
        require(tenant == msg.sender, "Only the tenant can make rent payments");
        require(msg.value == rentAmount, "Incorrect rent amount");

        rentPayments[msg.sender][_agreementId] += msg.value;

        emit RentPaid(msg.sender, _agreementId, msg.value);
    }

    function getRentPayments(address _tenant, uint256 _agreementId) public view returns (uint256) {
        return rentPayments[_tenant][_agreementId];
    }
}