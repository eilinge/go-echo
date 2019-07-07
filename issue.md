@[TOC](TCP/IP如何确保网络通讯质量)

# 操作步骤

## geth

    - geth --datadir ./data --networkid 15 --port 30303 --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcvhosts "*" --rpcapi "db,net,eth,web3,personal" --rpccorsdomain "*" --ws --wsaddr "localhost" --wsport "8546" --wsorigins "*" --nat "any" --nodiscover --dev --dev.period 1 console 2> 1.log

## mysqld server

    mysqld >> src/mysql.log

## mysql client

    - mysql -u root -p
    - 需要重新构建mysql数据库表

## run go

    go run main.go -c etc/copyright.dev.toml

# 其他

## net start mysql(服务名无效)

    https://www.jb51.net/article/51428.htm

## bytes32

    0xee440896028860593a2659daddb5817701dd9e04cc1e0f190120b1fe1e44d79a

### 需要注意的点

    1. 竞拍时, 是分布式, 需要最大值时, 需要使用互斥锁
    2. 拍卖时, 应该以weight所占比重, 计算金额?
    3. 排行榜, 选出投票的前10名(asset.voteCount), 分配一定erc20(基金会)
    4. 投票之后, 扣除一定的erc20给基金会
    5. 合约升级之后, 之前相关token/账户被存储到数据库中, 如何将数据存储到新的合约地址中?
    6. 用户刚刚注册进来,　没有ether就无法进行上传图片 -- 必须得用户自己本身拥有ether, 若是使用基金会直接转账, 则会导致多个账号进行注册, 将转出的ether进行买卖
    7. 使用node的express, 进行页面的布局(网页模板) -- 暂时无法使用(node(:3000)无法远程连接服务地址(http://localhost:8086))

### issues

    1. failed to instance.Mint no contract code at given address
        copyright.dev.toml合约地址错误
    2. failed to bind.NewTransactor could not decrypt key with given passphrase
        网页无法获取登录账户的address, 需要重新登录
    3. num, err := dbs.Create(sql), 不能使用num<=0做判断

### 数据库操作

    1. create table bidwinner (id int primary key not null auto_increment, token_id int not null unique, price int not null, address varchar(120));
    2. create table content (content_id int primary key not null auto_increment, title varchar(100), content varchar(256), content_hash varchar(100), price int, weight int,ts timestamp not null unique;

### 运行时遇到的问题

    1. 点击竞拍之后, 服务器未响应(提交的数据不对/服务器响应返回未被正确接收)
        将go eths.Auction 替换成 eths.Auction, 否则无法执行(if err!= nil{})
    2. 分割资产后, 需要对新, 旧资产进行更新, 存储content
        分割资产时, 添加emit newAsset()事件, 订阅到事件之后, 自动存储

### 待解决的问题

    1. 竞拍时, 资产拥有者无法进行拍卖
    2. 竞拍结束后, 假使未有人出到高于售价时的金额, 竞拍时间结束之后, 该资产应从竞拍列表中移除
    3. 查看所有资产的分页功能未完善(该项目还未完成请求vote?pageIndex=1的路由功能)
    4. echo 的模板 tempalte, 未找到更多的相关文档, 导致将网页与域名无法直接绑定
    ....

### 项目作业

    链下竞拍:
        新建bidwinner表, 将最高竞拍的资料进行存储, 竞拍结束时, 从bidwinner中, 提取出对应token, 完成资产分割和转账(自动:3minute)

    链下排行榜奖励:
        以第一次投票开始计时, 每10秒, 进行一次排行榜刷新
        10 * 10秒之后, 进行奖励

    erc20的获得和扣除:
        用户注册时, 会获得20(基金会)
        用户投票时, 会扣除30(基金会)
