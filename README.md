# README

## 任务要求：

用golang实现两个服务: api服务和算法服务，两个服务通过mq通信;

api服务提供3个接口:

1. 创建任务接口，创建一条任务插入mq，并返回taskid :

2. 结果查询接口，通过taskid查询redis任务完成状态放回;

3. 算法服务回调接口，当算法服务完成后通过该接口通知api服务，并将状态写入redis;

算法服务: 监听mq消费，通过cgo调用一个简单的C语言程序 (sleep3秒后输出hello word，模拟算法处理)，当收到结果后通过api的调接口上传处理状态

思考：

1. 当大并发请求时，如何提升服务可用性?

2. 耗时任务如何进行进度上报，取消任务等状态管理?

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

### RabbitMQ 工作流程

1、消息生产者连接到RabbitMQ Broker，创建connection，开启channel。

2、生产者声明交换机类型、名称、是否持久化等。

3、发送消息，并指定消息是否持久化等属性和routing key。

4、exchange收到消息之后，根据routing key路由到跟当前交换机绑定的相匹配的队列里面。

5、消费者监听接收到消息之后开始业务处理，然后发送一个ack确认告知消息已经被消费。

6、RabbitMQ Broker收到ack之后将对应的消息从队列里面删除掉。

### 开启图形界面插件

```
rabbitmq-plugins enable rabbitmq_management
```



## 完成记录

tag v0.1 实现了创建任务、查询任务、生产消息、消费消息、回调接口返回任务完成

tag v0.2 增加了状态回调、任务进度回调、取消任务

tag v0.2.2 修改了部分取消任务的bug，但是多个消费者时取消任务还是有bug，发布任务采用的时工作模式，但取消任务应该采取



## 思考:

### 问题一

当大并发请求时，可以通过以下方式提升服务可用性：

- 使用负载均衡：可以通过在 API 服务前面引入负载均衡器来均衡请求流量，避免单个 API 服务受到过多的请求。
- 增加服务器资源：可以增加 API 服务和算法服务的服务器数量，提高整个系统的并发处理能力。
- 优化算法处理程序：如果算法处理程序是整个系统的瓶颈，可以尝试优化算法处理程序，使其能够更快地完成任务处理。

### 问题二

1. 对于耗时任务，我们可以通过以下方式进行进度上报、取消任务等状态管理：

- 进度上报：耗时任务可以分为多个步骤，我们可以在每个步骤完成后，向消息队列或数据库中写入任务状态信息，例如当前处理进度、处理结果等，然后通过API接口提供查询任务状态的功能，让客户端可以随时查询任务的状态。
- 取消任务：当用户需要取消任务时，我们可以向任务状态信息中写入取消标记，并在任务处理过程中定时检查取消标记，如果检测到取消标记，则终止任务处理并将任务状态信息更新为取消状态。
- 状态管理：为了更好地管理任务状态，我们可以在Redis中创建一个任务状态管理系统，使用Redis的Hash数据类型存储任务状态信息，任务ID作为键，任务状态信息作为值，可以随时更新和查询任务状态。
