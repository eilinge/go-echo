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
