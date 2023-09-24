// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TenancyAgreement {
    struct Agreement {
        address landlord;
        address tenant;
        uint256 rentAmount;
        uint256 duration;
        uint256 startDate;
        bool isActive;
    }

    mapping (uint256 => Agreement) private agreements;
    uint256 private agreementCount;

    event AgreementCreated(uint256 agreementId);

    function createAgreement(address _landlord, address _tenant, uint256 _rentAmount, uint256 _duration, uint256 _startDate) public {
        agreements[agreementCount] = Agreement(_landlord, _tenant, _rentAmount, _duration, _startDate, true);
        emit AgreementCreated(agreementCount);
        agreementCount++;
    }

    function getAgreement(uint256 _agreementId) public view returns (address, address, uint256, uint256, uint256, bool) {
        Agreement memory agreement = agreements[_agreementId];
        return (agreement.landlord, agreement.tenant, agreement.rentAmount, agreement.duration, agreement.startDate, agreement.isActive);
    }

    function updateAgreementStatus(uint256 _agreementId, bool _isActive) public {
        agreements[_agreementId].isActive = _isActive;
    }
}