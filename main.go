package main

import (
	"fmt"

	"go-echo/configs"
	"go-echo/eths"
	"go-echo/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

var e *echo.Echo //echo框架对象全局定义
//静态html文件处理
func staticFile(e *echo.Echo) {

	e.Static("/", "static/pc/home") //根目录设置
	e.Static("/static", "static")   //全路径处理
	e.Static("/upload", "static/pc/upload")
	e.Static("/css", "static/pc/css")
	e.Static("/assets", "static/pc/assets")
	e.Static("/user", "static/pc/user")
}

func main() {

	fmt.Printf("get config %v ,%v\n", configs.Config.Common.Port, configs.Config.Db.Connstr)
	e = echo.New()             // 创建echo对象
	e.Use(middleware.Logger()) // 安装日志中间件
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	go eths.EventSubscribe("ws://localhost:8546", configs.Config.Eth.PxaAddr)

	staticFile(e) // 静态文件处理调用

	e.GET("/ping", routes.PingHandler)                  // 路由测试函数
	e.POST("/account", routes.Register)                 // register
	e.GET("/session", routes.GetSession)                // get session
	e.POST("/login", routes.Login)                      // login
	e.POST("/content", routes.UpLoad)                   // UpLoad
	e.GET("/content", routes.GetContents)               // GetPics
	e.GET("/content/:title", routes.GetContent)         // GetPic, /content/titleValues
	e.POST("/auction", routes.Auction)                  // 开始拍卖
	e.GET("/auctions", routes.GetAuctions)              // 查看拍卖
	e.GET("/auction/bid", routes.JoinBid)               // 参与拍卖
	e.GET("/vote", routes.VotePlace)                    // 投票中心
	e.GET("/content/vote", routes.Vote)                 // 进行投票
	e.Logger.Fatal(e.Start(configs.Config.Common.Port)) // 启动服务
}
