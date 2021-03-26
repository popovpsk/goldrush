package bank

import (
	"sync"
	"sync/atomic"
)

type Bank struct {
	coins []uint32
	l     sync.RWMutex
	idx   int32
	count int32
}

func NewBank() *Bank {
	b := &Bank{
		coins: make([]uint32, 0, 60000),
	}
	return b
}

func (b *Bank) Count() int32 {
	return atomic.LoadInt32(&b.count)
}

func (b *Bank) Store(coins []uint32) {
	if atomic.LoadInt32(&b.count) > 700 {
		return
	}
	b.l.Lock()
	b.coins = append(b.coins, coins...)
	b.l.Unlock()
	atomic.AddInt32(&b.count, int32(len(coins)))
}

func (b *Bank) Get(count int32) ([]uint32, bool) {
	//compare and shift count
	if c := atomic.LoadInt32(&b.count); c < count {
		return nil, false
	} else {
		for !atomic.CompareAndSwapInt32(&b.count, c, c-count) {
			c = atomic.LoadInt32(&b.count)
			if c < count {
				return nil, false
			}
		}
	}
	//ptr shift
	ptr := atomic.LoadInt32(&b.idx)
	for !atomic.CompareAndSwapInt32(&b.idx, ptr, ptr+count) {
		ptr = atomic.LoadInt32(&b.idx)
	}

	b.l.RLock()
	result := b.coins[ptr : ptr+count]
	b.l.RUnlock()

	return result, true
}
