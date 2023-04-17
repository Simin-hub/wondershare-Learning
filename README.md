# README
## redis的使用
[redis官网](https://redis.uptrace.dev/zh/guide/go-redis.html)
```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
	Addr:	  "localhost:6379",
	Password: "", // 没有密码，默认值
	DB:		  0,  // 默认DB 0
})
```
### 执行 Redis 命令
获取值
```go
val, err := rdb.Get(ctx, "key").Result()
fmt.Println(val)
```

## RabbitMQ

[参考地址](https://www.rabbitmq.com/tutorials/tutorial-one-go.html)

安装RabbitMQ

```
go get github.com/rabbitmq/amqp091-go
```

send.go

```go

```

## 最低完成度

tag v0.1
