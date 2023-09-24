pragma solidity ^0.8.0;

contract TenancyAgreement {
    enum AgreementStatus { Created, Accepted, Rejected, Deleted }

    struct Agreement {
        uint256 id;
        address landlord;
        address tenant;
        uint256 rentAmount;
        uint256 depositAmount;
        uint256 startDate;
        uint256 endDate;
        AgreementStatus status;
    }

    uint256 private agreementCounter;
    mapping(uint256 => Agreement) public agreements;

    event AgreementCreated(uint256 indexed id, address indexed landlord, address indexed tenant);
    event AgreementUpdated(uint256 indexed id, address indexed landlord, address indexed tenant);
    event AgreementDeleted(uint256 indexed id);
    event AgreementAccepted(uint256 indexed id, address indexed tenant);
    event AgreementRejected(uint256 indexed id, address indexed tenant);

    modifier onlyLandlord(uint256 id) {
        require(agreements[id].landlord == msg.sender, "Only landlord can perform this action");
        _;
    }

    modifier onlyTenant(uint256 id) {
        require(agreements[id].tenant == msg.sender, "Only tenant can perform this action");
        _;
    }

    function createAgreement(
        address tenant,
        uint256 rentAmount,
        uint256 depositAmount,
        uint256 startDate,
        uint256 endDate
    ) public {
        require(tenant != address(0), "Tenant address cannot be zero");
        require(rentAmount > 0, "Rent amount must be greater than zero");
        require(depositAmount >= 0, "Deposit amount must be non-negative");
        require(startDate < endDate, "Start date must be before end date");

        agreementCounter++;
        agreements[agreementCounter] = Agreement({
            id: agreementCounter,
            landlord: msg.sender,
            tenant: tenant,
            rentAmount: rentAmount,
            depositAmount: depositAmount,
            startDate: startDate,
            endDate: endDate,
            status: AgreementStatus.Created
        });

        emit AgreementCreated(agreementCounter, msg.sender, tenant);
    }

    function updateAgreement(
        uint256 id,
        uint256 rentAmount,
        uint256 depositAmount,
        uint256 startDate,
        uint256 endDate
    ) public onlyLandlord(id) {
        require(rentAmount > 0, "Rent amount must be greater than zero");
        require(depositAmount >= 0, "Deposit amount must be non-negative");
        require(startDate < endDate, "Start date must be before end date");
        require(agreements[id].status == AgreementStatus.Created, "Agreement cannot be updated");

        agreements[id].rentAmount = rentAmount;
        agreements[id].depositAmount = depositAmount;
        agreements[id].startDate = startDate;
        agreements[id].endDate = endDate;

        emit AgreementUpdated(id, msg.sender, agreements[id].tenant);
    }

    function deleteAgreement(uint256 id) public onlyLandlord(id) {
        require(agreements[id].status == AgreementStatus.Created, "Agreement cannot be deleted");

        delete agreements[id];
        emit AgreementDeleted(id);
    }

    function acceptAgreement(uint256 id) public onlyTenant(id) {
        require(agreements[id].status == AgreementStatus.Created, "Agreement cannot be accepted");

        agreements[id].status = AgreementStatus.Accepted;
        emit AgreementAccepted(id, msg.sender);
    }

    function rejectAgreement(uint256 id) public onlyTenant(id) {
        require(
            agreements[id].status == AgreementStatus.Created ||
            agreements[id].status == AgreementStatus.Accepted,
            "Agreement cannot be rejected"
        );

        agreements[id].status = AgreementStatus.Rejected;
        emit AgreementRejected(id, msg.sender);
    }
}