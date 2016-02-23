package queue

import (
	"container/list"
	"errors"
	"sync"
)

// BlockingQueue is a FIFO queue where Pop() operation is blocking if no items exists
type BlockingQueue struct {
	closed bool
	lock   sync.Mutex
	queue  *list.List

	notifyLock sync.Mutex
	monitor    *sync.Cond
}

// New instance of FIFO queue
func New() *BlockingQueue {
	bq := &BlockingQueue{queue: list.New()}
	bq.monitor = sync.NewCond(&bq.notifyLock)
	return bq
}

// Put any value to queue back. Returns false if queue closed
func (bq *BlockingQueue) Put(value interface{}) bool {
	if bq.closed {
		return false
	}
	bq.lock.Lock()
	defer func() {
		bq.notifyLock.Lock()
		bq.monitor.Signal()
		bq.notifyLock.Unlock()
	}()
	defer bq.lock.Unlock()
	bq.queue.PushBack(value)
	return true
}

// Pop front value from queue. Returns nil and false if queue closed
func (bq *BlockingQueue) Pop() (interface{}, bool) {
	if bq.closed {
		return nil, false
	}
	val, ok := bq.getUnblock()
	if ok {
		return val, ok
	}
	for !bq.closed {
		bq.notifyLock.Lock()
		bq.monitor.Wait()
		val, ok = bq.getUnblock()
		bq.notifyLock.Unlock()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

// Size of queue. Performance is O(1)
func (bq *BlockingQueue) Size() int {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	return bq.queue.Len()
}

// Closed flag
func (bq *BlockingQueue) Closed() bool {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	return bq.closed
}

// Close queue and explicitly remove each item from queue.
// Also notifies all reader (they will return nil and false)
// Returns error if queue already closed
func (bq *BlockingQueue) Close() error {
	if bq.closed {
		return errors.New("Already closed")
	}
	bq.closed = true
	bq.lock.Lock()
	//Clear
	for bq.queue.Len() > 0 {
		bq.queue.Remove(bq.queue.Front())
	}
	bq.lock.Unlock()
	bq.monitor.Broadcast()
	return nil
}

func (bq *BlockingQueue) getUnblock() (interface{}, bool) {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	if bq.queue.Len() > 0 {
		elem := bq.queue.Front()
		bq.queue.Remove(elem)
		return elem.Value, true
	}
	return nil, false
}
