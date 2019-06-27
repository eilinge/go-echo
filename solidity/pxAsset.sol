pragma solidity >=0.4.22 <0.6.0;

import "./ERC721.sol";
import "./pxcCoin.sol";
import "./ERC721TokenReceiver.sol";
import "./AddressUtils.sol";
import "./SafeMath.sol";

/*
weight: 财产分割
price: 拍卖
erc20  token: 注册/拍卖/转账/挖矿
erc721 token: 与财产进行绑定/挖矿/拍卖
*/

contract pxAsset is ERC721 {
    
    using AddressUtils for address;
    using SafeMath for uint;
    address public fundation;
    pxCoin pxcoin;
    bytes4 constant MAGIC_ON_ERC721_RECEIVED = 0x150b7a02;
    
    mapping(address=>uint) _ownerTokenCount; // owner=>cnts
    mapping(address=>uint[]) _ownerToken; // owner=>cnts[]
    mapping(uint=>address) _tokenOwner;//tokenId=>owner
    mapping(uint=>address) _tokenApprovals;//tokenId=>approved
    mapping(address=>mapping(address=>bool)) _operatorApprovals;//owner=>opertor=>bool
    
    event onNewAsset(bytes32 _hash, address _owner, uint _tokenId);
    
    struct Asset {
        bytes32 contentHash;
        uint price;
        uint weight;
        string metaData;
        uint voteCount;
    }
    
    Asset[] public assets;
    
    constructor() public {
        fundation = msg.sender;
        // assuer = msg.sender
        pxcoin = new pxCoin(1000000000, msg.sender);
    }
    
    function balanceOf(address _owner) external view returns (uint256){
        require( address(0) != _owner);
        return _ownerTokenCount[_owner];
    }
    
    function ownerOf(uint256 _tokenId) external view returns (address) {
        address tokenOwn = _tokenOwner[_tokenId];
        require( address(0) != tokenOwn);
        return tokenOwn;
    }
    
    modifier canTransfer(uint _tokenid) {
        address tokenOwner = _tokenOwner[_tokenid];
        require( msg.sender == tokenOwner || 
                 msg.sender == _getApproved(_tokenid) ||
                 _operatorApprovals[tokenOwner][msg.sender]);
        _;
    }
    modifier canOperate(uint _tokenid) {
        address tokenOwner = _tokenOwner[_tokenid];
        require( msg.sender == tokenOwner || _operatorApprovals[tokenOwner][msg.sender]);
        _;
    }
    modifier onlyOwner() {
        require( fundation == msg.sender);
        _;
    }
    modifier validToken( uint _tokenId) {
        address owner = _tokenOwner[_tokenId];
        require( owner != address(0) );
        _;
    }
    function _safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes _data) canTransfer(_tokenId) validToken(_tokenId) private {
        require( _to != address(0));
        _transfer(_from, _to, _tokenId);
        if (_to.isContract()) { // ?
            bytes4 retval = ERC721TokenReceiver(_to).onERC721Received(msg.sender, _from, _tokenId, _data);
            require( retval == MAGIC_ON_ERC721_RECEIVED);
        }
    }

    function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes _data) external payable {
        _safeTransferFrom(_from, _to, _tokenId, _data);
    }
    function safeTransferFrom(address _from, address _to, uint256 _tokenId) external payable {
        _safeTransferFrom(_from, _to, _tokenId, "");
    }
    
    function clearApproval( uint256 _tokenId) private {
        address owner = _tokenApprovals[_tokenId];
        require ( owner != address(0));
        delete _tokenApprovals[_tokenId];
    }
    
    function removeToken(address _from, uint256 _tokenId) private {
        require ( _tokenOwner[_tokenId] == _from);
        assert (_ownerTokenCount[_from] > 0);
        _ownerTokenCount[_from]--;
        delete _tokenOwner[_tokenId];
    }
    
    function addToken(address _to, uint256 _tokenId) private {
        require( _tokenOwner[_tokenId] == address(0));
        _tokenOwner[_tokenId] = _to;
        _ownerTokenCount[_to]++;
    }
    
    function _transfer( address _from, address _to, uint256 _tokenId) private {
        address owner = _tokenOwner[_tokenId];
        require(owner == _from);
        // clear approve
        clearApproval(_tokenId);
        // change tokenOwner
        removeToken( owner, _tokenId);
        addToken(_to, _tokenId);
        emit Transfer(owner, _to, _tokenId);
    }
    function transferFrom(address _from, address _to, uint256 _tokenId) canTransfer(_tokenId) validToken(_tokenId) external payable { // _from ?
        // address owner = _tokenOwner[_tokenId];
        require( _to != address(0));
        _transfer(_from, _to, _tokenId);
    }
    function approve(address _approved, uint256 _tokenId) canOperate(_tokenId) validToken(_tokenId) external payable {
        // address tokenOwner = _tokenOwner[_tokenId];
        require( _approved != address(0));
        _tokenApprovals[_tokenId] = _approved;
    }
    function setApprovalForAll(address _operator, bool _approved) external {
        require( _operator != address(0));
        require( _ownerTokenCount[msg.sender] > 0);
        _operatorApprovals[msg.sender][_operator] = _approved;
    }
    function getApproved(uint256 _tokenId) external view returns (address) {
        return _getApproved(_tokenId);
    }
    function _getApproved(uint256 _tokenId) canOperate(_tokenId) private view returns (address) {
        return _tokenApprovals[_tokenId];
    }
    function isApprovedForAll(address _owner, address _operator) external view returns (bool) {
        return _operatorApprovals[_owner][_operator];
    }
    
    function _newAsset(bytes32 _contentHash, uint _price, uint _weight, string _data) private returns(uint) {
        Asset memory a = Asset(_contentHash, _price, _weight, _data, 0);
        // 添加新的资产, 并获得token_id
        uint num = assets.push(a) - 1;
        return num;
    }
    
    function mint(bytes32 _contentHash, uint _price, uint _weight, string _data) external {
        uint tokenId = _newAsset(_contentHash, _price, _weight, _data);
        _ownerTokenCount[msg.sender] = _ownerTokenCount[msg.sender].add(1);
        _tokenOwner[tokenId] = msg.sender;
        emit onNewAsset(_contentHash, msg.sender, tokenId);
        pxcoin.transfer(msg.sender, 100);
        _ownerToken[msg.sender].push(tokenId);
    }
    
    function _approveAsset(address _approved, uint256 _tokenId) canOperate(_tokenId) validToken(_tokenId) private {
        require( _approved != address(0));
        _tokenApprovals[_tokenId] = _approved;
    }

    function auction(address _from, uint _price, uint _tokenId, address _to) public {
        _approveAsset(_to, _tokenId);
        _transfer(_from, _to, _tokenId);
        pxcoin.approve(_to, _price);
        pxcoin.transferFrom(_from, _to, _price);
    }
    
    function splitAsset(uint _tokenId, uint _weight, address _buyer) onlyOwner() canTransfer(_tokenId) validToken(_tokenId) public returns(uint) {
        require(_weight <= 100);
        require( address(0) != _buyer);
        Asset storage a = assets[_tokenId];
        require(a.weight >= _weight);
        
        uint256 tokenId = assets.push(a) - 1;
        Asset storage newA = assets[tokenId];
        a.weight = a.weight.sub(_weight);
        newA.weight = _weight;
        newA.voteCount = uint(0);
        addToken(_buyer, tokenId);
        _ownerToken[_buyer].push(tokenId);
        return tokenId;
    }
    
    function vote(uint _tokenId) public {
        require(_tokenId < assets.length);
        Asset storage a = assets[_tokenId];
        a.voteCount = a.voteCount.add(1);
        pxcoin.transfer(_tokenOwner[_tokenId], 10);
    }
    
    function getPXCBalance(address _owner) view public returns(uint256) {
        return pxcoin.balanceOf(_owner);
    }
    
    function getPXCAddr() view public returns(address) {
        return address(pxcoin);
    }
    
    function getOwnerToken(address _owner) public view returns(uint[]) {
        require( address(_owner) != 0);
        return _ownerToken[_owner];
    }
}
