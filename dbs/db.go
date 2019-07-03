package dbs

import (
	"database/sql"
	"fmt"
	"go-echo/configs"
	_ "strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// content: content->content_hash
// account_content: content_hash -> token_id ->address

const defaultFormat = "2006-01-02 15:04:05 PM"

type (
	// Accounts ...
	Accounts struct {
		Email       string `json:"email"`
		IdentitiyID string `json:"identity_id"`
		UserName    string `json:"username"`
	}
	// Content ...
	Content struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		ContentHash string `json:"content_hash"`
		Price       int64  `json:"price"`
		Weight      int64  `json:"weight"`
	}
	// Auction ...
	Auction struct {
		ContentHash string `json:"content_hash"`
		Address     string `json:"address"`
		TokenID     int64  `json:"token_id"`
		Percent     int64  `json:"percent"`
		Price       int64  `json:"price"`
	}
	// BidPerson ...
	BidPerson struct {
		TokenID int64  `json:"token_id"`
		Price   int64  `json:"price"`
		Address string `json:"address"`
	}
)

// DBConn 数据库连接的全局变量
var DBConn *sql.DB

//init 函数是本包被其他文件引用时自动执行，并且整个工程只会执行一次
func init() {
	fmt.Println("call dbs.Init", configs.Config)
	DBConn = InitDB(configs.Config.Db.Connstr, configs.Config.Db.Driver)
}

// InitDB  初始化数据库连接
func InitDB(connstr, Driver string) *sql.DB {
	db, err := sql.Open(Driver, connstr) // connect mysql
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return db
}

// DBQuery 通用查询，返回map嵌套map
func DBQuery(sql string) ([]map[string]string, int, error) {
	// fmt.Println("query is called:", sql)
	rows, err := DBConn.Query(sql)
	if err != nil {
		fmt.Println("query data err", err)
		return nil, 0, err
	}
	//得到列名数组
	cols, err := rows.Columns()
	//获取列的个数
	colCount := len(cols)
	values := make([]string, colCount)
	oneRows := make([]interface{}, colCount)
	for k, _ := range values {
		oneRows[k] = &values[k] //将查询结果的返回地址绑定，这样才能变参获取数据
	}
	//存储最终结果
	results := make([]map[string]string, 1)
	idx := 0
	//循环处理结果集
	for rows.Next() {
		rows.Scan(oneRows...)
		rowmap := make(map[string]string)
		for k, v := range values {
			rowmap[cols[k]] = v

		}
		if idx > 0 {
			results = append(results, rowmap)
		} else {
			results[0] = rowmap
		}
		idx++
		//fmt.Println(values)
	}
	//fmt.Println("---------------------------------------")

	return results, idx, nil

}

// Create ....create/delete/update
func Create(sql string) (int64, error) {
	res, err := DBConn.Exec(sql)
	if err != nil {
		fmt.Println("exec sql err,", err, "sql is ", sql)
		return -1, err
	}
	return res.LastInsertId()
}

// AddContent ...
func (c *Content) AddContent() error {
	// ts := time.Now().Format(defaultFormat)
	b := []byte(time.Now().Format(defaultFormat))
	ts := string(b[:len(b)-3])
	// ts := strings.Replace(time.Now().Format(defaultFormat), " PM", "", 1)
	sql := fmt.Sprintf("insert into content(title, content, content_hash, price, weight, ts) values('%s', '%s', '%s', '%d', '%d','%s')",
		c.Title, c.Content, c.ContentHash, c.Price, c.Weight, ts)
	fmt.Println("insert into content", sql)
	res, err := DBConn.Exec(sql)
	if err != nil {
		fmt.Println("failed to insert content ", err)
		return err
	}
	_, err = res.LastInsertId()
	if err != nil {
		fmt.Println("failed to res.LastInsertId ", err)
		return err
	}
	// fmt.Println("id==", id)
	return nil
}
