package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func Add(a, b uint) uint {
	return a + b
}

func Sub(a, b uint) uint {
	return a - b
}

func pointStream(intPoint *int) *int {
	*intPoint = *intPoint + 10
	return intPoint
}

func slicePoint(p *[]int) {
	for i := 0; i < len(*p); i++ {
		(*p)[i] = (*p)[i] * 2
		fmt.Println((*p)[i])
	}
}

func goroutineUse() {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			if i%2 == 1 {
				fmt.Println("goroutine1>>>", i)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			if i%2 == 0 {
				fmt.Println("goroutine2>>>", i)
			}
		}
	}()

	wg.Wait()
}

// 任务模型
type Task struct {
	ID   int
	Name string
}

func (t Task) New(i int, name string) Task {
	return Task{i, name}
}

func (t Task) prt() (time.Duration, error) {
	now := time.Now()
	delay := 10 + rand.Intn(10)
	time.Sleep(time.Millisecond * time.Duration(delay))
	deltaTime := time.Since(now)
	return deltaTime, nil
}

// 任务执行结果模型
type TaskResult struct {
	TaskID   int
	TaskName string
	Duration time.Duration
	Error    error
}

// 任务调度器模型
type TaskScheduler struct {
	//任务列表
	tasks []Task
	//最大并发数
	maxConcurrency int
	//并发信号量
	semaphore chan struct{}
	//并发访问互斥锁
	mutex sync.Mutex
	//任务结果列表
	results []TaskResult
	//并发同步
	wg sync.WaitGroup
}

func NewTaskScheduler(maxConcurrency int) *TaskScheduler {
	if maxConcurrency <= 0 {
		maxConcurrency = 3
	}
	return &TaskScheduler{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

func (ts *TaskScheduler) Add(task Task, id int, name string) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.tasks = append(ts.tasks, task.New(id, name))
}

func (ts *TaskScheduler) Run() {
	//并发接收结果
	//resultChain := make(chan TaskResult, len(ts.tasks))
	ts.results = make([]TaskResult, len(ts.tasks))

	for index, task := range ts.tasks {
		ts.wg.Add(1)
		go func(t Task, i int) {
			defer ts.wg.Done()

			//使用结构体类型通道，最主要是0内存占用，且通道是线程安全的，做到了内存屏障
			ts.semaphore <- struct{}{}
			defer func() { <-ts.semaphore }()

			duration, err := t.prt()
			tr := TaskResult{
				TaskID:   t.ID,
				TaskName: t.Name,
				Duration: duration,
				Error:    err,
			}

			ts.mutex.Lock()
			ts.results[i] = tr
			ts.mutex.Unlock()

			////结果推送到通道里面去，安全
			//resultChain <- tr

		}(task, index)
	}
	ts.wg.Wait()
	//close(resultChain)

	////确保同步写入数据，避免竟态异常，原因是切片动态扩容，除非已知结果，提前预设切片长度，这样性能最高
	//for tr := range resultChain {
	//	ts.mutex.Lock()
	//	ts.results = append(ts.results, tr)
	//	ts.mutex.Unlock()
	//}
}

func (ts *TaskScheduler) Statistics() []string {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	fmt.Println("实际完成任务数：", len(ts.results), len(ts.tasks))

	results := make([]string, len(ts.results))
	for index, tr := range ts.results {
		results[index] = strconv.Itoa(tr.TaskID) + "-" + tr.TaskName + "-" + tr.Duration.String()
		fmt.Println(results[index])
	}
	return results
}

func TestSchedular() {
	ts := NewTaskScheduler(5)
	taskNum := 10
	for i := 0; i < taskNum; i++ {
		ts.Add(Task{}, i, "task:"+strconv.Itoa(i))
	}
	fmt.Println(ts.tasks)
	ts.Run()
	ts.Statistics()
}

/*
题目 ：定义一个 Shape 接口，包含 Area() 和 Perimeter() 两个方法。
然后创建 Rectangle 和 Circle 结构体，实现 Shape 接口。
在主函数中，创建这两个结构体的实例，并调用它们的 Area() 和 Perimeter() 方法。
*/

type Shape interface {
	Area() float64
	Perimeter() float64
	Name() string
}

type Rectangle struct {
	X, Y float64
	n    string
}

func (r Rectangle) Area() float64 {
	return r.X * r.Y
}

func (r Rectangle) Perimeter() float64 {
	return 2*r.X + 2*r.Y
}

func (r Rectangle) Name() string {
	return r.n
}

type Circle struct {
	R float64
	n string
}

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.R
}

func (c Circle) Name() string {
	return c.n
}

/*
使用组合的方式创建一个 Person 结构体，包含 Name 和 Age 字段，
再创建一个 Employee 结构体，组合 Person 结构体并添加 EmployeeID 字段。
为 Employee 结构体实现一个 PrintInfo() 方法，输出员工的信息
*/
type Person struct {
	Name string
	Age  int
}

type Employee struct {
	EmployeeID int
	Person     Person
}

func (e Employee) PrintInfo() {
	fmt.Println(e.EmployeeID, e.Person.Name, e.Person.Age)
}

/*
编写一个程序，使用通道实现两个协程之间的通信。
一个协程生成从1到10的整数，并将这些整数发送到通道中，
另一个协程从通道中接收这些整数并打印出来。
考察点 ：通道的基本使用、协程间通信
*/
func ChannelWithoutCache() {
	ch := make(chan int)
	//保证进程执行
	var wg *sync.WaitGroup
	wg = &sync.WaitGroup{}
	num := 10
	wg.Add(1)
	go func() {
		defer wg.Done()
		//数据生产完毕，要关闭通道，避免内存泄漏风险
		defer close(ch)
		for i := 1; i <= num; i++ {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(110)))
			ch <- i
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		maxRetries := 3
		attempts := 0

		for {
			for i := 1; i <= maxRetries; i++ {
				//设置定时器100ms
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
				defer cancel()
				select {
				case v, ok := <-ch:
					if !ok {
						println("通道已关闭")
						return
					}
					fmt.Println(v)
				case <-ctx.Done():
					attempts++
					fmt.Printf("第%d次检查，暂无数据\n", attempts)
					if attempts >= maxRetries {
						fmt.Printf("第%d次检查，无数据,退出\n", attempts)
						return
					}
				}
			}

		}
	}()

	wg.Wait()
}

/*
实现一个带有缓冲的通道，
生产者协程向通道中发送100个整数，
消费者协程从通道中接收这些整数并打印。
考察点 ：通道的缓冲机制。
*/
func ChannelWithCache() {
	ch := make(chan int, 3)
	wg := &sync.WaitGroup{}
	num := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		for i := 1; i <= num; i++ {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			ch <- i
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		maxRetries := 3
		attempts := 0
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()
			select {
			case v, ok := <-ch:
				if !ok {
					fmt.Println("通道已关闭")
					return
				}
				fmt.Println(v)
			case <-ctx.Done():
				attempts++
				fmt.Printf("第%d等待数据\n", attempts)
				if attempts >= maxRetries {
					fmt.Printf("第%d等待数据，退出\n", attempts)
					return
				}
			}
		}
	}()
	wg.Wait()
}

/*
题目 ：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。
启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
考察点 ： sync.Mutex 的使用、并发数据安全。
*/
func BlockLock() {
	t := time.Now()
	count := 0
	coroutineNum := 10
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	addNum := 10000

	for i := 0; i < coroutineNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < addNum; n++ {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	deltaTime := time.Since(t)
	fmt.Printf("耗时：%v; 结果：%v\n", deltaTime, count)
}

/*
题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。
启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
考察点 ：原子操作、并发数据安全。
*/
func DataLock() {
	var count int64
	t := time.Now()
	count = 0
	coroutineNum := 10
	wg := &sync.WaitGroup{}
	addNum := 10000
	for i := 0; i < coroutineNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < addNum; n++ {
				atomic.AddInt64(&count, 1)
			}
		}()
	}
	wg.Wait()
	deltaTime := time.Since(t)
	fmt.Printf("耗时：%v; 结果：%v\n", deltaTime, count)
}

func main() {
	//i := 10
	//fmt.Println(&i)
	//r := pointStream(&i)
	//fmt.Println(r)
	//fmt.Println(*r)
	//fmt.Println(i)
	//p := []int{1, 2, 3, 4, 5}
	//slicePoint(&p)
	//goroutineUse()
	//TestSchedular()

	//shapeList := []Shape{
	//	Rectangle{1, 2, "Rectangle"},
	//	Circle{5, "Circle"},
	//}
	//
	//for _, shape := range shapeList {
	//	fmt.Println(shape.Name(), shape.Area(), shape.Perimeter())
	//}
	//e := Employee{
	//	EmployeeID: 1,
	//	Person:     Person{"tom", 18},
	//}
	//e.PrintInfo()
	//ChannelWithoutCache()
	//ChannelWithCache()
	//BlockLock()
	DataLock()

}
