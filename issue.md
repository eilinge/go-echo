# Issue

## net start mysql(服务名无效)

    https://www.jb51.net/article/51428.htm

## geth

    geth --datadir ./data --networkid 15 --port 30303 --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcvhosts "*" --rpcapi "db,net,eth,web3,personal" --rpccorsdomain "*" --ws --wsaddr "localhost" --wsport "8546" --wsorigins "*" --nat "any" --nodiscover --dev --dev.period 1 console 2> 1.log

## mysqld server

    mysqld >> src/mysql.log

## mysql client

    mysql -u root -p

## run go

    go run main.go -c etc/copyright.dev.toml

## bytes32

    0xee440896028860593a2659daddb5817701dd9e04cc1e0f190120b1fe1e44d79a

### 需要注意的点

    1. 竞拍时, 是分布式, 需要最大值时, 需要使用互斥锁
    2. 拍卖时, 应该以weight所占比重, 计算金额
    3. 排行榜, 选出投票的前10名(asset.voteCount), 分配一定erc20(基金会)
    4. 投票之后, 扣除一定的erc20给基金会
    5. 合约升级之后, 之前相关token/账户被存储到数据库中, 如何将数据存储到新的合约地址中
    6. 用户刚刚注册进来,　没有ether就无法进行上传图片
    7. 使用node的express, 进行页面的布局(网页模板)

### issues

    1. failed to instance.Mint no contract code at given address
        copyright.dev.toml合约地址错误

    2. failed to bind.NewTransactor could not decrypt key with given passphrase
        网页无法获取登录账户的address, 需要重新登录

    3. num, err := dbs.Create(sql)
        不能使用num<=0做判断

    4. 刚注册的用户如果没有ether, 进行图片认证, 则无法进行上传图片认证, 但是会被存储新的资产到content数据库中
        先进行图片认证, 当无法进行图片认证之后, 返回错误报告, 如果成功, 再进行存储content操作

    5. 进行资产分割时, 假使用户erc721 token不足, 也会先进行资产分割成功, 无法完成转账操作
        先对用户资产余额判断操作, 假使不足, 返回错误报告, 如果成功, 进行资产分割成功, 和转账操作

### 数据库操作

    1. create table bidwinner (id int primary key not null auto_increment, token_id int not null unique, price int not null, address varchar(120));
    2. create table content (content_id int primary key not null auto_increment, title varchar(100), content varchar(256), content_hash varchar(100) unique, price int, weight int,ts timestamp not null;

### 程序逻辑

    1. 当2个操作没有直接关联时, 将数据库操作, 转账操作放到后面执行, 提前对其不足条件进行判断
