package account

import "sync"

type transactionQueue struct {
	customerID identifier
	queue      []*loadTransaction
	mutex      *sync.RWMutex
}

func newTransactionQueue(customerID identifier) *transactionQueue {
	return &transactionQueue{
		customerID: customerID,
		queue:      make([]*loadTransaction, 0, 5),
		mutex:      &sync.RWMutex{},
	}
}

func (q *transactionQueue) pushBack(t *loadTransaction) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queue = append(q.queue, t)
}

func (q *transactionQueue) popFront() *loadTransaction {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.queue) == 0 {
		return nil
	}
	t := q.queue[0]
	q.queue = q.queue[1:]

	return t
}

func (q *transactionQueue) isEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.queue) == 0
}
