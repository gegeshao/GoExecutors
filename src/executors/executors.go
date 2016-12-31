package executors

import (
	"config"
	"fmt"
	"sync"
	"time"
)

type Callable struct {
	f *func() interface{}
}
type CallableQ chan Callable

func NewCallable(f *func() interface{}) *Callable {
	return &Callable{f}
}

type Executors struct {
	callableQ CallableQ
	lock      *sync.Mutex
	rets      map[Callable]chan interface{}
}

type Future struct {
	es       *Executors
	callable Callable
}

func NewFuture(es *Executors, callable Callable) *Future {
	return &Future{es, callable}
}

func (f *Future) GetResult(timeout time.Duration) interface{} {
	fmt.Println("future get result.callable:", f.callable)
	timer := time.NewTimer(timeout)
	var ret interface{}
	defer delete(f.es.rets, f.callable)
	fmt.Println("future对应的结果chan:", f.es.rets[f.callable])
	select {
	case ret = <-f.es.rets[f.callable]:
		fmt.Println("future 获取到了结果：", ret)
		return ret
	case <-timer.C:
		return nil
	}

}

func NewExecutors() *Executors {
	var cq = make(CallableQ, 100)
	var es = &Executors{cq, &sync.Mutex{}, make(map[Callable]chan interface{})}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			var callable Callable
			for {
				callable = <-cq

				var ret = (*callable.f)()
				fmt.Println("callable:", callable, " ret:", ret, " to ", es.rets[callable])
				es.lock.Lock()
				defer es.lock.Unlock()
				var retChan = es.rets[callable]
				fmt.Println("result chan:", retChan)
				retChan <- ret
			}
		}()
	}

	return es

}

func (es *Executors) Submit(callable Callable) *Future {
	es.callableQ <- callable
	es.rets[callable] = make(chan interface{}, 1) // 保证map里面有东西
	return NewFuture(es, callable)
}
