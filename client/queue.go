package main

import (
	"os"
	"sync"
)

type Queue struct {
	queue chan *Args
	wg    *sync.WaitGroup
}

type Args struct {
	file  *os.File
	index int
	size  int64
	n     chan int64
}

func newQueue() *Queue {
	return &Queue{
		queue: make(chan *Args, 1024),
		wg:    new(sync.WaitGroup),
	}
}

func (q *Queue) start(num int, deal func(*Args)) {
	for i := 0; i < num; i++ {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			for args := range q.queue {
				deal(args)
			}
		}()
	}
}

func (q *Queue) stop() {
	close(q.queue)
	q.wg.Wait()
}

func (q *Queue) submit(a *Args) {
	q.queue <- a
}
