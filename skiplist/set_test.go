package skiplist

import (
	"testing"
)

type value struct {
	score uint64
	key   uint64
}

func (this *value) Key() uint64   { return this.key }
func (this *value) Score() uint64 { return this.score }
func (this *value) ReCalcScore()  {}

type cmp struct {
}

func (this *cmp) CmpScore(v1 interface{}, v2 interface{}) int {
	s1 := v1.(*value).score
	s2 := v2.(*value).score
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}

func (this *cmp) CmpKey(v1 interface{}, v2 interface{}) int {
	s1 := v1.(*value).key
	s2 := v2.(*value).key
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}

func Test(t *testing.T) {

	ss := NewSet(&cmp{})
	key_1 := &value{
		score: 10,
		key:   1,
	}
	ss.Insert(key_1)
	key_2 := &value{
		score: 10,
		key:   2,
	}
	ss.Insert(key_2)
	key_3 := &value{
		score: 8,
		key:   3,
	}
	ss.Insert(key_3)
	key_4 := &value{
		score: 11,
		key:   4,
	}
	ss.Insert(key_4)
	ss.Dump()

	println("Key 3, rank: ", ss.GetRank(3))

	rg := &RangeSpec{
		Min: 10,
		Max: 10,
	}
	println("Delete Rank:", ss.DeleteRangeByScore(rg))
	ss.Dump()

	keys := ss.GetRange(1, 15, false)
	for _, val := range keys {
		println("key: ", val.(*value).score, val.(*value).Key())
	}

	ss.DeleteRangeByRank(1, 3)
	ss.Dump()
}
