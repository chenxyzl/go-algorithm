package skiplist

import (
	"errors"
	"math/rand"
	"testing"
)

type value struct {
	key   uint64
	score uint64
}

func (this *value) Key() uint64   { return this.key }
func (this *value) Score() uint64 { return this.score }

type cmp struct {
}

func (this *cmp) CmpScore(v1 *value, v2 *value) int {
	s1 := v1.score
	s2 := v2.score
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}

func (this *cmp) CmpKey(v1 *value, v2 *value) int {
	s1 := v1.key
	s2 := v2.key
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}

func checkSort(sl *SkipList[*value], t *testing.T) {
	min := uint64(0)
	sl.Range(func(v *value) bool {
		if v.Score() < min {
			t.Error()
		}
		min = v.Score()
		return true
	})
}

func TestCRUD(t *testing.T) {
	//	ss := NewSet(&cmp{})
	//  set := make(map[uint64] *value)
	sl := NewSkipList[*value](&cmp{})
	count := uint64(100)
	for i := uint64(0); i < count; i++ {
		key := &value{
			score: uint64(rand.Uint32()%100 + 1),
			key:   i,
		}
		sl.Insert(key)
	}
	//长度
	if sl.Length() != int(count) {
		t.Error()
	}
	//遍历
	checkSort(sl, t)
	//增加
	v := &value{
		key:   count / 2,
		score: count / 2,
	}
	_, err := sl.Insert(v)
	if !errors.Is(err, RepeatedKey) {
		t.Error()
	}
	//长度
	if sl.Length() != int(count) {
		t.Error()
	}

	//增加
	{
		v = &value{
			key:   count,
			score: count / 2,
		}
		_, err = sl.Insert(v)
		if err != nil {
			t.Error()
		}
		//长度
		if sl.Length() != int(count+1) {
			t.Error()
		}
		checkSort(sl, t)
	}
	//删除
	{
		has := sl.Delete(50)
		if has != 1 {
			t.Error()
		}
		has = sl.Delete(50)
		if has != 0 {
			t.Error()
		}
		//长度
		if sl.Length() != int(count) {
			t.Error()
		}
		checkSort(sl, t)
	}
}
