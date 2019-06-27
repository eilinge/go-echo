pragma solidity >=0.4.22 <0.6.0;

/**
 * @dev Utility library of inline functions on addresses.
 */
library AddressUtils {

  /**
   * @dev Returns whether the target address is a contract.
   * @param _addr Address to check.
   */
  function isContract(
    address _addr
  )
    internal
    view
    returns (bool)
  {
    uint256 size;

    /**
     * XXX Currently there is no better way to check if there is a contract in an address than to
     * check the size of the code at that address.
     * See https://ethereum.stackexchange.com/a/14016/36603 for more details about how this works.
     * TODO: Check this again before the Serenity release, because all addresses will be
     * contracts then.
     */

    //如何查看一个地址是EOS还是智能合约地址，就在于查看该账户代码的长度。EOS下是不能存放代码的，所以它的长度是0，只有合约账户的地址是能够存放代码的
    assembly { size := extcodesize(_addr) } // solium-disable-line security/no-inline-assembly
    return size > 0;
  }

}