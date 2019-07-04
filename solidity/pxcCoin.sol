pragma solidity >=0.4.22 <0.6.0;
import "./ERC20.sol";


contract pxCoin is ERC20 {
    string public name = "pxcb";
    string public symbol = "pxc";
    
    address public fundation;
    address public assuer;
    uint private _totalSupply;
    
    mapping(address=>uint) _balance;
    mapping(address=>mapping(address=>uint)) _allowance; // 授权余额
    
    constructor(uint totalSupply, address _owner) public {
        _totalSupply = totalSupply;
        fundation = _owner;
        assuer = msg.sender;
        _balance[fundation] = totalSupply * 20 / 100;
        _balance[assuer] = totalSupply * 80 / 100;
    }
    
    function totalSupply() public view returns (uint totalsupply) { // 获取总的发行量
        totalsupply = _totalSupply;
        return;
    }

    function balanceOf(address _owner) public view returns (uint balance){ // 查询账户余额
        balance = _balance[_owner];
        return;
    }
    function transfer(address _to, uint _value) public returns(bool success){ // 发送Token到某个地址(转账)
        if (_balance[msg.sender] >= _value && _balance[_to]+_value > 0) {
            _balance[msg.sender] -= _value;
            _balance[_to] += _value;
            emit Transfer(msg.sender, _to, _value);
            return true;
        } else {
            return false;
        }
    }
    
    function transferFrom(address _from, address _to, uint _value) public returns (bool success){ // 从地址from 发送token到to地址
        if (_balance[_from] >= _value && _balance[_to]+_value > 0 && address(0) != _to && _allowance[_from][_to] >= _value) {
            _balance[_from] -= _value;
            _balance[_to] += _value;
            _allowance[_from][_to] -= _value;
            emit Transfer(_from, _to, _value);
            return true;
        } else {
            return false;
        }
    }
    
    function approve(address _spender, uint _value) public returns(bool success){ // 允许_spender从你的账户转出token
        if (_balance[msg.sender] >= _value && address(0) != _spender) {
            _allowance[msg.sender][_spender] = _value;
            emit Approval(msg.sender, _spender, _value);
            return true;
        } else {
            return false;
        }
    }
    
    function allowance(address _owner, address _spender) public view returns (uint remaining) { // 询允许spender转移的Token数量
        remaining = _allowance[_owner][_spender];
        return;
    }
    
    function getAddress() public view returns(address) {
        return address(this);
    }
    
}