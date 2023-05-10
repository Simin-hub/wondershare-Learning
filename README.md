# README

[项目地址](https://github.com/Simin-hub/wondershare-Learning)

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

进阶问题：

1. 试着在C.Hello()种写入一段随机导致程序崩溃的代码（让程序崩溃，不要捕获），看看会发生什么；这种情况怎么处理？
2. 算法任务的处理时间和任务有长有短，怎么处理才能保证大部分用户的等待时间可控？
3. 代码结构模块化分层

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

tag v0.2.2 修改了部分取消任务的bug，但是多个消费者时取消任务还是有bug，发布任务采用的工作模式，但取消任务应该采取

tag v0.2.3 修改了部分取消任务的bug，在更新任务状态和进度时，查询任务是否取消，若取消则返回错误。

tag v0.2.4 增加了协程池和任务超时设置，调整代码结构层次

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

## 进阶问题

### 问题一：

参考：[CGO编程](https://chai2010.cn/advanced-go-programming-book/ch2-cgo/index.html)、[在GO中调用C源代码](https://www.cnblogs.com/oscar2960/p/16242310.html)、[golang 捕获 C/C++ 错误并做善后处理](https://blog.csdn.net/u013272009/article/details/113617745)、[CGO编程？其实没有想象的那么难！](https://mdnice.com/writing/2ea753c3d2cd4101b11761900e3fbb95)、

CGO 在使用 C/C++资源的时候一般有三种形式：

1. 直接使用源码；
2. 链接静态库；
3. 链接动态库。

在C.Hello()中写入一段随机导致程序崩溃的代码，会导致算法服务进程崩溃。cgo 中的 crash ，在 golang 中是捕获不到信号量的，如信号量SIGSEGV。在 Go 代码中使用 `cgo` 调用 C 函数时，如果 C 代码中存在未处理的错误，可能会导致程序崩溃，而这种崩溃通常是由于操作系统发送了一个信号量（如 `SIGSEGV`）导致的。在这种情况下，Go 代码无法捕获信号量并继续执行，因为信号量是由操作系统直接发送到进程中的，Go 运行时无法完全控制它们的行为。在 CGO 中调用 C 程序时，如果 C 程序崩溃，Go 程序也会崩溃，可能会对系统的稳定性和可靠性造成严重影响。

解决办法:

1. 为了解决这个问题，可以在算法服务进程启动时使用进程监控工具，如supervisord，来监控算法服务进程的运行状态，并在进程崩溃时自动重启服务进程。
2. 在 C 代码中使用错误处理机制来捕获和处理异常。例如，可以使用 C 语言中的 `setjmp()` 和 `longjmp()` 函数实现非局部跳转，从而在异常发生时跳转到指定的处理逻辑。另外，可以使用 C 语言中的异常处理机制，如 `try-catch` 语句块，来捕获和处理异常。
3. 隔离CGO，把使用 cgo 模块单独拎出来做成无状态服务，使用它的服务与它放在同一个 pod 中，使用 unix socket 通信。



### 问题二：

要保证大部分用户的等待时间可控，可以采用一些优化策略。例如，可以通过增加算法服务的处理能力，来缩短任务处理时间。另外，可以对任务进行分批处理，同时处理多个任务，从而提高服务的并发处理能力，减少用户的等待时间。还可以通过优化算法算法实现，缩短任务的处理时间，提高算法服务的响应速度。 

具体的几个解决办法：

1. 将任务分为不同的优先级，根据优先级进行调度。处理时间短且紧急的任务会被优先处理，而处理时间长且非紧急的任务则可以被延迟处理。
2. 设定一个最大任务处理时间，若超出限制则返回算法处理错误。
3. 可以使用线程池来处理任务。线程池可以在程序启动时创建一定数量的线程，每个线程可以并发处理多个任务。当一个任务完成时，线程可以从任务队列中取出下一个任务进行处理。这样可以让算法任务在一定程度上并发执行，减少等待时间。
4. 对于处理时间较长的任务，可以考虑采用分布式计算的方式来加速处理。将一个大的任务分成多个小的子任务，分别在不同的机器上进行处理，最后再将结果合并起来。
5. 启用多个算法服务，在使用Kubernetes等容器化技术部署算法服务时，可以使用自动伸缩功能。自动伸缩可以根据当前任务请求的负载情况自动调整算法服务的副本数量，从而提高系统的弹性和响应速度。
6. 算法优化，例如采用并行计算、分布式计算等方式，提高算法服务处理速度。
7. 状态监控，在任务处理过程中，可以实时监控任务的状态，包括任务的等待时间、处理时间、执行进度等信息，通过实时监控可以及时发现任务处理速度慢的情况，进行调整和优化，保证任务的及时处理和响应速度。

实现了两种方式：任务处理超时以及线程池的方式

设置单个任务的执行最长时间

```
	// 创建一个带有超时的context
	child, cancelFunc := context.WithTimeout(ctx, time.Second*50)
	go func() {
		for i := 1; i <= 10; i++ {
			// 分段任务时，每次都检查任务是否超时
			if child.Err() != nil {
				task.Status = TASK_TIMEDOUT
				break
			}

			// 单次处理任务
			StartProcess(task.ID)

			task.Process = strconv.Itoa(i * 10)
			err := UpdateTask(task.ID, TASK_PROCESSING, task.Process)
			if err != nil {
				// updating the task process fails
				log.Printf("The task %s has been cancelled", task.ID)
				break
			}
			log.Printf("Task id:%s\t process:%s\t", task.ID, task.Process)
		}
		// 任务完成
		cancelFunc()
	}()

	// 等待任务超时或者完成
	select {
	case <-child.Done():
	}
```

设置一个线程池，从消息队列读取消息后则根据有缓冲的chan创建协程

```
// 创建一个任务数量通道, MaxWokers 最大协程数量
	taskQueue := make(chan int, MaxWokers)
	var wg sync.WaitGroup
	for d := range msgs {
		// decode message
		task := &Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("Failed to decode task message: ", err)
			continue
		} else {
			log.Println("Ongoing consumption task: ", task.ID)
		}
		// 从消息队列中获取任务，向chan申请协程，若chan已满则等待
		taskQueue <- 1
		go func() {
			defer wg.Done()
			wg.Add(1)
			worker(ctx, task)
			<-taskQueue
		}()
	}
	// 等待所有的协程完成
	wg.Wait()
```

### 问题三：

参考：[Go 面向包的设计和架构分层（cmd, internal, pkg)](https://studygolang.com/articles/30164)

将程序分为了四个模块，分别是算法服务（`algorithm`）、API 服务（`api`）和程序入口（cmd）、公共包（pkg）。

下面是代码结构模块化分层

<img src="https://raw.githubusercontent.com/Simin-hub/Picture/master/img/image-20230510161618083.png" alt="image-20230510161618083" style="zoom: 80%;" />

`cmd`模块是程序的入口，包含两个程序入口： 算法服务程序入口（`algorithm.go`）、API服务程序入口（`api.go` ）。

`internal` 模块下的所有包及相应文件都有一个项目保护级别，即其他项目是不能导入这些包的，仅仅是该项目内部使用。`internal/algorithm/` 存放的是算法服务程序的代码，`cgo/` 存放使用cgo和c的代码、算法服务代码（`alg.go`）。`internal/api/` 是存放API服务相关代码，api服务代码（`service.go` ），api请求处理（`handler.go` ）。`internal/pkg/`是项目中的程序都可以分享的代码， 自定义错误代码（`errors.go`），定义的模型代码（`Task.go` ）。

这种结构可以很好地实现模块化分层，每个模块负责不同的功能，且彼此之间相互独立，易于维护和扩展。
