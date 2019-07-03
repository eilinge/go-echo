package routes

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"go-echo/configs"
	"go-echo/dbs"
	"go-echo/eths"
	"go-echo/utils"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

// PageMaxPic ...
const PageMaxPic = 5

var mutex sync.Mutex

const defaultFormat = "2006-01-02 15:04:05 PM"

// Price ...
var Price int64

// PingHandler ...
func PingHandler(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

/*
17601329166@163.com
eilinge
eilinge
*/

var passwd string

// Register ...
func Register(c echo.Context) error {
	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//2. 解析数据
	account := &dbs.Accounts{}

	/*
		将前端传过来的数据, 与dbs.Accounts进行数据绑定
		&dbs.Account{
			Email       `json:"email"`			name="email"
			IdentitiyID `json:"identity_id"`	name="identity_id"
			UserName 	`json:"username"`		name="username"
		}
	*/
	if err := c.Bind(account); err != nil { // 解析form表单
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
	//3. 操作geth创建账户(account.IdentitiyId->pass)
	passwd = account.IdentitiyID
	address, err := eths.NewAcc(account.IdentitiyID, configs.Config.Eth.Connstr)
	if err != nil {
		// fmt.Println("failed to NewAcc: ", err)
		resp.Errno = utils.RECODE_IPCERR
		return err
	}
	go func() {
		err = eths.EthErc20Transfer(configs.Config.Eth.Fundation, configs.Config.Eth.FundationPWD, address, 5)
		if err != nil {
			fmt.Println("Transfer failed when register err: ", err)
			return
		}
		// _, err = eths.EtherTransfer(configs.Config.Eth.Fundation, address)
	}()
	//4. 操作Mysql插入数据
	pwd := fmt.Sprintf("%x", sha256.Sum256([]byte(account.IdentitiyID)))
	sql := fmt.Sprintf("insert into account(email, username, identity_id, address) values('%s', '%s', '%s', '%s')",
		account.Email, account.UserName, pwd, address)
	fmt.Println(sql)
	_, err = dbs.Create(sql)
	if err != nil {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	//5. session处理
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	sess.Values["address"] = address
	sess.Values["username"] = account.UserName
	sess.Save(c.Request(), c.Response())
	return nil
}

// Login ...
func Login(c echo.Context) error {
	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//2. 解析数据
	account := &dbs.Accounts{}

	if err := c.Bind(account); err != nil { // 解析form表单
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
	fmt.Println("account.IdentitiyID: ", account.UserName, account.IdentitiyID)
	passwd = account.IdentitiyID

	//3. 操作Mysql查询数据
	pwd := fmt.Sprintf("%x", sha256.Sum256([]byte(account.IdentitiyID)))
	sql := fmt.Sprintf("select * from account where username='%s' and identity_id='%s';",
		account.UserName, pwd)
	// fmt.Println(sql)
	values, num, err := dbs.DBQuery(sql)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DATAERR
		return err
	}
	row1 := values[0]
	//5. session处理
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	sess.Values["address"] = row1["address"]
	sess.Values["username"] = row1["username"]
	sess.Save(c.Request(), c.Response())
	return nil
}

// GetSession ....
func GetSession(c echo.Context) error {
	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address := sess.Values["address"]
	// username := sess.Values["username"]
	if address == nil {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	return nil
}

// UpLoad ...
func UpLoad(c echo.Context) error {
	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//2. 解析数据
	content := &dbs.Content{}

	price, _ := strconv.ParseInt(c.FormValue("price"), 10, 32)
	weight, _ := strconv.ParseInt(c.FormValue("weight"), 10, 32)

	content.Price = price
	content.Weight = weight

	h, err := c.FormFile("fileName") // 解析文件名
	if err != nil {
		fmt.Println("failed to FormFile: ", err)
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
	src, err := h.Open()
	defer src.Close()
	// 3. 打开一个目标文件用于保存
	content.Content = "static/photo/" + h.Filename
	dst, err := os.Create(content.Content)
	if err != nil {
		fmt.Println("failed to create file: ", err)
		resp.Errno = utils.RECODE_IOERR
		return err
	}
	defer dst.Close()

	// 4. get hash
	cData := make([]byte, h.Size)
	n, err := src.Read(cData)
	if err != nil || h.Size != int64(n) {
		resp.Errno = utils.RECODE_IOERR
		return err
	}
	content.ContentHash = fmt.Sprintf("%x", sha256.Sum256(cData))
	dst.Write(cData) // 图片存储

	content.Title = h.Filename
	// 5. write to dbs / 给上传图片页面, 添加weight, price, 并一起传入
	// content.AddContent()

	// 6. 操作以太坊
	sess, _ := session.Get("session", c)
	fromAddr, ok := sess.Values["address"].(string)
	if fromAddr == "" || !ok {
		resp.Errno = utils.RECODE_SESSIONERR
		return errors.New("no session")
	}
	// from, pass, hash, data string
	fmt.Printf("price: %d, weight: %d\n", price, weight)
	go func() {
		err = eths.Upload(fromAddr, passwd, content.ContentHash, content.Title, price, weight)
		if err != nil {
			resp.Errno = utils.RECODE_IPCERR
			return
		}
		content.AddContent()
	}()
	return nil
}

// GetContents 根据用户查找出其所有资产
func GetContents(c echo.Context) error {
	time.Sleep(time.Second * 3)

	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address := sess.Values["address"]
	if address == nil {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	sql := fmt.Sprintf("select a.content_hash,weight,a.title,b.token_id from content a, account_content b where a.content_hash = b.content_hash and address='%s'", address)
	fmt.Println(sql)
	contents, num, err := dbs.DBQuery(sql)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	totalPage := int(num)/PageMaxPic + 1
	currentPage := 1
	mapResp := make(map[string]interface{})
	mapResp["current_page"] = currentPage
	mapResp["total_page"] = totalPage
	mapResp["contents"] = contents

	resp.Data = mapResp
	return nil
}

// GetContent ...
func GetContent(c echo.Context) error {
	//1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	content := &dbs.Content{}
	content.Title = c.Param("title")
	// fmt.Println("get title: ", content.Title)
	// 2. 查询数据库获得文件路径
	sql := fmt.Sprintf("select * from content where title='%s'", content.Title)
	value, num, err := dbs.DBQuery(sql)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	content.Content = value[0]["content"]
	http.ServeFile(c.Response(), c.Request(), content.Content)
	return nil
}

// Auction ...
func Auction(c echo.Context) error {
	// 1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	// 2. 获取session 的address
	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address, ok := sess.Values["address"].(string)
	if address == "" || !ok {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}

	// 3. 解析参数
	auction := &dbs.Auction{}
	fmt.Printf("start parse from ......................")
	if err := c.Bind(auction); err != nil { // 解析form表单, string -> int64
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}

	if auction.Percent <= 0 || auction.Price <= 0 {
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
	// 4. 插入拍卖(auction)数据库
	b := []byte(time.Now().Format(defaultFormat))
	ts := string(b[:len(b)-3])
	auction.Address = address
	sql := fmt.Sprintf("insert into auction(content_hash, address, token_id, percent, price, status, ts) value('%s','%s', %d, %d, %d, 1,'%s')",
		auction.ContentHash, auction.Address, auction.TokenID, auction.Percent, auction.Price, ts)
	fmt.Println(sql)
	_, err = dbs.Create(sql)
	if err != nil {
		resp.Errno = utils.RECODE_DBERR
		fmt.Println("failed to dbs.Create(sql)...")
		return err
	}
	fmt.Println("start insert into bidWinner...")
	// 4.5 插入bidwinner数据库
	Price = auction.Price
	theWinner := &dbs.BidPerson{Price: auction.Price, Address: auction.Address}
	b1 := []byte(time.Now().Format(defaultFormat))
	ts1 := string(b1[:len(b1)-3])
	WinnerSQL := fmt.Sprintf("insert into bidwinner(token_id, price, address, ts) values('%d', '%d','%s', '%s');", auction.TokenID, theWinner.Price, theWinner.Address, ts1)
	fmt.Println("winner: ", WinnerSQL)
	_, err = dbs.Create(WinnerSQL)
	if err != nil {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	fmt.Println("--------------------------------------")
	fmt.Println("start bid, this asset 2 minute over")
	// 5. 开始拍卖执行后, 设置定时器, 时间结束, 自动完成财产的分割/erc20转账
	ticker := time.NewTicker(time.Minute * 2)
	go func() {
		for i := 1; i > 0; i-- {
			// for {
			<-ticker.C
			EndBid(auction.TokenID, auction.Percent)
		}
	}()
	return nil
}

// GetAuctions ...
func GetAuctions(c echo.Context) error {
	time.Sleep(time.Second)
	// 1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	// 2. 获取session 的address
	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address, ok := sess.Values["address"].(string)
	if address == "" || !ok {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}

	// 3. 查看拍卖
	// 自动识别出查询字段所在tables
	// sql := fmt.Sprintf("select a.*,b.title from auction a, content b where a.content_hash=b.content_hash and a.status=1;")
	sql := fmt.Sprintf("select a.content_hash,title,b.price,b.percent,token_id from content a, auction b where a.content_hash=b.content_hash and b.status=1")
	fmt.Println(sql)
	values, num, err := dbs.DBQuery(sql)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	mapResp := make(map[string]interface{})
	fmt.Printf("the values: %#v\n", values)
	mapResp["data"] = values

	resp.Data = mapResp
	return nil
}

// JoinBid ...
func JoinBid(c echo.Context) error {
	// 1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	// 2. 获取参数
	// price,tokenid
	price := c.QueryParam("price")
	tokenID := c.QueryParam("tokenid")

	// 3. session
	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address, ok := sess.Values["address"].(string)
	if address == "" || !ok {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	// 进行比较竞拍值, 保存该address
	_price, _ := strconv.ParseInt(price, 10, 64)
	_tokenid, _ := strconv.ParseInt(tokenID, 10, 64)

	// 从bidwinner中取出当前最大的price, 然后进行比较
	maxSQL := fmt.Sprintf("select price from bidwinner where token_id='%d'", _tokenid)
	fmt.Println(maxSQL)
	value, num, err := dbs.DBQuery(maxSQL)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	Price, _ = strconv.ParseInt(value[0]["price"], 10, 64)

	// 同步锁, 防止多人同时修改数据
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Printf("the price: %d and Price: %d\n", _price, Price)
	if _price > Price {
		Price = _price
		theWinner := &dbs.BidPerson{Price: _price, Address: address}
		sql := fmt.Sprintf("update bidwinner set price = '%d', address ='%s' where token_id='%d';", theWinner.Price, theWinner.Address, _tokenid)
		fmt.Println("update bidwinner: ", sql)
		_, err := dbs.Create(sql)
		if err != nil {
			resp.Errno = utils.RECODE_DBERR
			return err
		}
		fmt.Printf("the account: %s Join bid success ...", address)
	}
	return nil
}

// EndBid ...
func EndBid(tokenID, weight int64) error {
	// 3.5 根据tokenId, 查询出最高价者的address, price
	WinSQL := fmt.Sprintf("select token_id,price,address from bidwinner where token_id='%d'", tokenID)
	winDetail, num, err := dbs.DBQuery(WinSQL)
	if err != nil || num <= 0 {
		fmt.Println("failed to WinSQL...", err)
		return err
	}
	price, _ := strconv.ParseInt(winDetail[0]["price"], 10, 32)
	address := winDetail[0]["address"]
	// 4. 数据库操作, price
	// 4.1 获取拍卖时的价格
	priAuctSQL := fmt.Sprintf("select price from  auction where token_id='%d' and status=1", tokenID)
	priceAuct, num, err := dbs.DBQuery(priAuctSQL)
	if err != nil || num <= 0 {
		fmt.Println("failed to priAuctSQL select...", err)
		return err
	}
	_priceAuct := priceAuct[0]["price"]

	sql := fmt.Sprintf("update auction set price='%d',status=0 where token_id='%d'", price, tokenID)
	_, err = dbs.Create(sql)
	if err != nil {
		fmt.Println("failed to update auction...", err)
		return err
	}
	// 4.1 更新content数据库: percent(总的-参与拍卖的)/price()
	weightSQL := fmt.Sprintf("select a.percent,b.weight,b.price,a.content_hash from auction a, content b where a.content_hash = b.content_hash and token_id ='%d'", tokenID)
	auctionWeight, num, err := dbs.DBQuery(weightSQL)
	if err != nil || num <= 0 {
		fmt.Println("failed to weightSQL...", err)
		return err
	}
	// aPercent, _ := strconv.ParseInt(auctionWeight[0]["percent"], 10, 32)
	bWeight, _ := strconv.ParseInt(auctionWeight[0]["weight"], 10, 32)
	bPrice, _ := strconv.ParseInt(auctionWeight[0]["price"], 10, 32)
	contentHash := auctionWeight[0]["content_hash"]
	_newPriceAuct, _ := strconv.ParseInt(_priceAuct, 10, 32)
	// 总的权重  - 拍卖时的权重
	newWeight := bWeight - weight
	// 拍卖之后的价格 - 拍卖时的价格 + 原本的价格
	newPrice := (price - _newPriceAuct) + bPrice
	// content_hash 不唯一(分割出的新资产与旧资产content_hash)
	UpConSQL := fmt.Sprintf("update content set price='%d' ,weight='%d' where content_hash ='%s';", newPrice, newWeight, contentHash)
	_, err = dbs.Create(UpConSQL)
	if err != nil {
		fmt.Println("failed to update content...", err)
		return err
	}
	// 获取token_id最高竞拍者的price
	bidSQL := fmt.Sprintf("select price,address from auction where token_id = '%d'", tokenID)
	value, num, err := dbs.DBQuery(bidSQL)
	if err != nil || num <= 0 {
		fmt.Println("failed to bidSQL...", err)
		return err
	}
	to := value[0]["address"]

	fmt.Println("---------------------------------------")
	fmt.Println("aleady EndBid, Waiting SpiltAsset and transfer .....")
	// 5. 操作以太坊: 资产分割, erc20转账
	go func() {
		// 提前判断后面情况的条件, 保证转账成功: 余额 > 参与拍卖的价格
		balance, err := eths.GetPxcBalance(address)
		if err != nil || balance < price {
			fmt.Println("your pxa balance less")
			return
		}
		err = eths.EthSplitAsset(configs.Config.Eth.Fundation, configs.Config.Eth.FundationPWD, address, tokenID, weight)
		if err != nil {
			fmt.Println("failed to eths.EthSplitAsset ", err)
			return
		}
		// _price, _ := strconv.ParseInt(price, 10, 32)
		err = eths.EthErc20Transfer(address, configs.Config.Eth.FundationPWD, to, price)
		if err != nil {
			fmt.Println("failed to eths.EEthErc20Transfer ", err)
			return
		}
		fmt.Println("---------------------------------------")
		fmt.Println("Success SpiltAsset and transfer .....")
	}()
	return nil
}

// Vote ...
func Vote(c echo.Context) error {
	// 1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	// 2. 获取参数
	tokenID, _ := strconv.ParseInt(c.QueryParam("token_id"), 10, 32)
	contentHash := c.QueryParam("contentHash")
	// 3. session
	sess, err := session.Get("session", c)
	if err != nil {
		fmt.Println("failed to get session")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	address, ok := sess.Values["address"].(string)
	if address == "" || !ok {
		fmt.Println("failed to get address")
		resp.Errno = utils.RECODE_SESSIONERR
		return err
	}
	// 3.5 获取该address响应erc20 余额, 保证其有足够(>=30)的token进行该次投票
	erc20Balance, _ := eths.GetPxcBalance(address)
	if erc20Balance < 30 || err != nil {
		fmt.Println("your erc20 balance is poor, connot operate this vote")
		return err
	}
	// 4. 存储到数据库
	b := time.Now().Format(defaultFormat)
	ts := b[:len(b)-3]
	VoteSQL := fmt.Sprintf("insert into vote(address, token_id, vote_time, comment) value('%s', '%d', '%s', '%s')", address, tokenID, ts, contentHash)
	fmt.Println("VoteSQL: ", VoteSQL)
	_, err = dbs.Create(VoteSQL)
	if err != nil {
		fmt.Println("failed to VoteSQL")
		resp.Errno = utils.RECODE_DATAERR
		return err
	}
	// 5. 操作以太坊, 进行投票, 只能合约内将erc20 token转给tokenID的地址
	eths.VoteTo(address, configs.Config.Eth.FundationPWD, tokenID)
	return nil
}
