package bank

import "testing"

func TestBank(t *testing.T) {
	b := NewBank()

	st1 := int32(321)
	st2 := int32(25432)

	sl := make([]uint32, 0, st1)
	for i := 0; i < int(st1); i++ {
		sl = append(sl, uint32(i))
	}
	b.Store(sl)

	sl = make([]uint32, 0, st2)
	for i := 0; i < int(st2); i++ {
		sl = append(sl, uint32(i))
	}
	b.Store(sl)

	if b.Count() != st1+st2 {
		t.Errorf("cnt=%v != %v", b.Count(), st1+st2)
	}

	c := []uint32{9990, 9991, 9992, 9993}

	b.Store(c)

	res, ok := b.Get(4)
	if !ok || len(res) != 4 {
		t.Error()
	}
	if b.Count() != st1+st2 {
		t.Errorf("cnt=%v != %v", b.Count(), st1+st2)
	}

	c2 := []uint32{7770, 7771, 7772, 7773}
	b.Store(c2)

	res2, ok2 := b.Get(8)
	if !ok2 || len(res2) != 8 {
		t.Error()
	}
}
