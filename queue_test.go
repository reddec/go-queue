package queue

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreatePutQueue(t *testing.T) {
	queue := New()
	if !queue.Put(0) {
		t.Fatal("Any value must be accepted in new queue")
	}
}

func TestFillReadOneThread(t *testing.T) {
	queue := New()
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i
		if !queue.Put(i) {
			t.Fatal("Any value must be accepted in new queue")
		}
	}
	//Check
	nsum := 0
	for queue.Size() > 0 {
		t.Log(queue.Size())
		val, ok := queue.Pop()
		if !ok {
			t.Fatal("Not all values available")
		}
		nsum += val.(int)
	}
	if nsum != sum {
		t.Fatal("Not all items received")
	}
}

func TestFillReadManyThreads(t *testing.T) {
	queue := New()
	nsum := int32(0)

	for i := 0; i < 12; i++ {
		go func(thread int) {
			for {
				v, ok := queue.Pop()
				if !ok {
					break
				}
				atomic.AddInt32(&nsum, v.(int32))
				fmt.Println("Thread ", thread)
			}
		}(i)
	}

	sum := 0
	for i := 0; i < 40; i++ {
		sum += i
		if !queue.Put(int32(i)) {
			t.Fatal("Any value must be accepted in new queue")
		}
	}

	time.Sleep(100 * time.Millisecond)
	if err := queue.Close(); err != nil {
		t.Fatal("Queue must close", err)
	}
	if nsum != int32(sum) {
		t.Fatal("Not all items received")
	}
}
