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

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

// PageMaxPic ...
const PageMaxPic = 5

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
	fmt.Println("start prase pragma...")
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
	fmt.Println("start prase pragma...")
	if err := c.Bind(account); err != nil { // 解析form表单
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
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
	// 5. write to dbs
	content.AddContent()

	// 6. 操作以太坊
	sess, _ := session.Get("session", c)
	fromAddr, ok := sess.Values["address"].(string)
	if fromAddr == "" || !ok {
		resp.Errno = utils.RECODE_SESSIONERR
		return errors.New("no session")
	}
	// from, pass, hash, data string
	go eths.Upload(fromAddr, passwd, content.ContentHash, content.Title)
	return nil
}

// GetContents 根据用户查找出其所有资产
func GetContents(c echo.Context) error {
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
	sql := fmt.Sprintf("select a.content_hash,a.title,b.token_id from content a, account_content b where a.content_hash = b.content_hash and address='%s'", address)
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
	fmt.Println("get title: ", content.Title)
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
	if err := c.Bind(auction); err != nil { // 解析form表单, string -> int64
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}

	if auction.Percent <= 0 || auction.Price <= 0 {
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}
	// 4. 插入拍卖数据库
	auction.Address = address
	sql := fmt.Sprintf("insert into auction(content_hash, address, token_id, percent, price, status) value('%s','%s',%d,%d,%d,1)",
		auction.ContentHash, auction.Address, auction.TokenID, auction.Percent, auction.Price)
	fmt.Println(sql)
	num, err := dbs.Create(sql)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	return nil
}

// GetAuctions ...
func GetAuctions(c echo.Context) error {
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
	sql := fmt.Sprintf("select a.content_hash,title,price,token_id from content a, auction b where a.content_hash=b.content_hash and b.status=1")
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

// EndBid ...
func EndBid(c echo.Context) error {
	// 1. 响应数据结构初始化
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	// 2. 获取参数
	// weight,tokenid
	weight := c.QueryParam("weight")
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

	// 4. 数据库操作
	sql := fmt.Sprintf("update auction set percent='%s',status=0 where token_id='%s'", weight, tokenID)
	_, err = dbs.Create(sql)
	if err != nil {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	bidSQL := fmt.Sprintf("select price,address from auction where token_id = '%s'", tokenID)
	value, num, err := dbs.DBQuery(bidSQL)
	if err != nil || num <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	to := value[0]["address"]
	price := value[0]["price"]

	// 5. 操作以太坊: 资产分割, erc20转账
	go func() {
		_tokenID, _ := strconv.ParseInt(tokenID, 10, 32)
		_weight, _ := strconv.ParseInt(weight, 10, 32)
		err = eths.EthSplitAsset(configs.Config.Eth.Fundation, "eilinge", address, _tokenID, _weight)
		if err != nil {
			fmt.Println("failed to eths.EthSplitAsset ", err)
			return
		}
		_price, _ := strconv.ParseInt(price, 10, 32)
		err = eths.EthErc20Transfer(address, "eilinge", to, _price)
		if err != nil {
			fmt.Println("failed to eths.EEthErc20Transfer ", err)
			return
		}
	}()
	return nil
}
