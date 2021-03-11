package game

import (
	"sync"
)

type Bank struct {
	coins []uint32
	l     sync.Mutex
}

func NewBank() *Bank {
	return &Bank{
		coins: make([]uint32, 0, 600000),
	}
}

func (b *Bank) Count() int {
	return len(b.coins)
}

func (b *Bank) Store(coins []uint32) {
	b.l.Lock()
	b.coins = append(b.coins, coins...)
	b.l.Unlock()
}

func (b *Bank) Get(count int, result []uint32) bool {
	b.l.Lock()
	defer b.l.Unlock()
	if len(b.coins) < count {
		return false
	}
	sl := b.coins[len(b.coins)-count:]
	b.coins = b.coins[:len(b.coins)-count]
	copy(result, sl)
	return true
}
