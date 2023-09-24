// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./TenancyAgreement.sol";

contract SecurityDeposit {
    TenancyAgreement private tenancyAgreement;
    mapping (address => mapping (uint256 => uint256)) private deposits;

    constructor(address _tenancyAgreementAddress) {
        tenancyAgreement = TenancyAgreement(_tenancyAgreementAddress);
    }

    event DepositPaid(address indexed tenant, uint256 agreementId, uint256 amount);
    event DepositReturned(address indexed tenant, uint256 agreementId, uint256 amount);

    function makeDeposit(uint256 _agreementId) public payable {
        (, address tenant, , , , ) = tenancyAgreement.getAgreement(_agreementId);
        require(tenant == msg.sender, "Only the tenant can make a security deposit");

        deposits[msg.sender][_agreementId] += msg.value;

        emit DepositPaid(msg.sender, _agreementId, msg.value);
    }

    function getDeposit(address _tenant, uint256 _agreementId) public view returns (uint256) {
        return deposits[_tenant][_agreementId];
    }

    function returnDeposit(uint256 _agreementId) public {
        (address landlord, address tenant, uint256 rentAmount, uint256 duration, uint256 startDate, bool isActive) = tenancyAgreement.getAgreement(_agreementId);
        require(!isActive, "Tenancy agreement is still active");
        require(landlord == msg.sender, "Only the landlord can return the security deposit");

        uint256 depositAmount = deposits[tenant][_agreementId];
        deposits[tenant][_agreementId] = 0;
        payable(tenant).transfer(depositAmount);

        emit DepositReturned(tenant, _agreementId, depositAmount);
    }
}