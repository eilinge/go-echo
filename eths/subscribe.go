package eths

import (
	"context"
	"errors"
	"fmt"
	"go-echo/dbs"
	"reflect"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
0x
12345678000000000000000000000000   - hash
00000000000000000000000000000000
0000000000000000000000003bf47cd6   - address
d8587374b891545fa1f08e196bd6c135
00000000000000000000000000000000
00000000000000000000000000000000   - tokenID

0x
ee440896028860593a2659daddb58177
01dd9e04cc1e0f190120b1fe1e44d79a
000000000000000000000000d64d314d
d56fc742f85f8ef77ba9abf58b26c867
00000000000000000000000000000000
00000000000000000000000000000004

0x
759963ee59d46872e1798e0ccdfaf788
d737302f2712089ad3e6067b075c9168
00000000000000000000000088c1ce3a
a95955c49dbf54a070dc8e8611cbaaca
00000000000000000000000000000000
00000000000000000000000000000001
*/

const defaultFormat = "2006-01-02 15:04:05 PM"

// LogDataUnpack ...
func LogDataUnpack(start, end int, val interface{}, data []byte) (err error) {
	length := len(data)
	fmt.Println("call--- LogDataUnpack begin", reflect.TypeOf(val).String(), length)

	if start >= length || end > length {
		return errors.New("too short datas")
	}
	pdata := data[start:end]

	fmt.Println(string(data), string(pdata))
	if reflect.TypeOf(val).String() == "int64" || reflect.TypeOf(val).String() == "*int64" {
		var tmpval *int64 = val.(*int64)
		// string -> int64 16: 0x, 32: int32
		*tmpval, err = strconv.ParseInt(string(pdata), 16, 32)
		fmt.Println("call ParseInt", val)
	} else if reflect.TypeOf(val).String() == "string" || reflect.TypeOf(val).String() == "*string" {
		var tmpval *string = val.(*string)
		*tmpval = string(pdata)
		fmt.Println("call ParseInt", val)
	}

	fmt.Println("call--- LogDataUnpack end", val)
	return nil
}

// ParseMintEventDb1 ...
func ParseMintEventDb1(data []byte) error {
	fmt.Println(string(data))
	var tokenID int64
	err := LogDataUnpack(32*5, 32*6, &tokenID, data)
	if err != nil {
		fmt.Println("faile to get tokenID", err)
		return err
	}
	fmt.Println("tokenID===", tokenID)
	var pixHash string
	err = LogDataUnpack(32*0, 32*2, &pixHash, data)
	if err != nil {
		return err
	}
	fmt.Println("pixHash===", pixHash)
	var pixAddr string
	err = LogDataUnpack(32*3-8, 32*4, &pixAddr, data)
	if err != nil {
		return err
	}
	pixAddr = "0x" + pixAddr
	fmt.Println("pixAddr===", pixAddr)
	//插入到数据库中
	b := []byte(time.Now().Format(defaultFormat))
	ts := string(b[:len(b)-3])
	sql := fmt.Sprintf("insert into account_content(content_hash,token_id,address) values('%s',%d,'%s', '%s')", pixHash, tokenID, pixAddr, ts)
	_, err = dbs.Create(sql)
	if err != nil {
		fmt.Println("failed to insert into mysql ", sql, err)
		return err
	}
	return nil
}

// ParseMintEventDb ...
func ParseMintEventDb(data []byte) error {
	fmt.Println(string(data))
	var tokenID int64
	tokenID, _ = strconv.ParseInt(string(data[32*5:32*6]), 16, 32)
	fmt.Println("tokenID===", tokenID)

	var pixHash string
	pixHash = string(data[0 : 32*2])
	fmt.Println("pixHash===", pixHash)

	var pixAddr string
	pixAddr = string(data[32*3-8 : 32*4])
	pixAddr = "0x" + pixAddr
	fmt.Println("pixAddr===", pixAddr)
	// 插入到用户资产(account_content)数据库中
	sql := fmt.Sprintf("insert into account_content(content_hash,token_id,address) values('%s',%d,'%s')", pixHash, tokenID, pixAddr)
	_, err := dbs.Create(sql)
	if err != nil {
		fmt.Println("failed to insert into mysql ", sql, err)
		return err
	}
	return nil
}

// EventSubscribe ...
func EventSubscribe(connstr, contractAddr string) error {
	cli, err := ethclient.Dial(connstr)
	if err != nil {
		fmt.Println("failed to connet to geth", err)
		return err
	}
	// 合约地址处理
	cAddress := common.HexToAddress(contractAddr)
	newAssetHash := crypto.Keccak256Hash([]byte("onNewAsset(bytes32,address,uint256)"))
	// 过滤处理
	query := ethereum.FilterQuery{
		Addresses: []common.Address{cAddress},
		Topics:    [][]common.Hash{{newAssetHash}},
	}
	// 通道
	logs := make(chan types.Log)
	// 订阅
	sub, err := cli.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		fmt.Println("failed to SubscribeFilterLogs", err)
		return err
	}
	// 订阅返回处理
	fmt.Println("start operate sub response ...")
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("get sub err", err)
		case vLog := <-logs:
			data, err := vLog.MarshalJSON()
			fmt.Println("sub response: ", string(data), err)
			ParseMintEventDb([]byte(common.Bytes2Hex(vLog.Data)))
		}
	}
}

/*
string(data):
{
	"address":"0x7c4aea600f4cf70a79784c20b6bf17b57e873354",
	"topics":["0x656ff40b8e1f65ac976060cda3f8f3e146ac6bb0421c1fe1010c92c350e57db3"],
	"data":"0x586ea5dd46b08604d74905a6583008f3dabd57eec7ba3915177cbd90baa5364000000000000000000000000088c1ce3aa95955c49dbf54a070dc8e8611cbaaca0000000000000000000000000000000000000000000000000000000000000003",
	"blockNumber":"0x148b2",
	"transactionHash":"0xfec0722b4f17bf8a44cf547ef3d7406c800afd2704fbe7f19595c2e6933be248",
	"transactionIndex":"0x0",
	"blockHash":"0xffb4e90919072f21691208888178b0f3bb9d1fb269e13f8a87ca2f46874c7a80",
	"logIndex":"0x0",
	"removed":false
}
*/
