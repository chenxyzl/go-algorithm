package skiplist

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
)

var P float32 = 0.25

const maxLevel uint32 = 32

var RepeatedKey = errors.New("RepeatedKey")

type Valuer interface {
	Key() uint64
	Score() uint64
}

type Level[V Valuer] struct {
	forward *Node[V]
	span    uint32
}

type Node[V Valuer] struct {
	value    V
	backward *Node[V]
	level    []Level[V]
}

func NewNode[V Valuer](level uint32, value V) *Node[V] {
	sln := &Node[V]{
		value: value,
		level: make([]Level[V], level),
	}
	return sln
}

func (n *Node[V]) Next(i int) *Node[V] {
	return n.level[i].forward
}

func (n *Node[V]) SetNext(i int, next *Node[V]) {
	n.level[i].forward = next
}

func (n *Node[V]) Span(i int) uint32 {
	return n.level[i].span
}

func (n *Node[V]) SetSpan(i int, span uint32) {
	n.level[i].span = span
}

func (n *Node[V]) Value() V {
	return n.value
}

func (n *Node[V]) Prev() *Node[V] {
	return n.backward
}

type Comparator[V Valuer] interface {
	CmpScore(V, V) int
	CmpKey(V, V) int
}

type SkipList[V Valuer] struct {
	head, tail *Node[V]
	level      uint32
	length     int
	cmp        Comparator[V]
	index      map[uint64]V
}

func NewSkipList[V Valuer](cmp Comparator[V]) *SkipList[V] {
	sl := &SkipList[V]{
		level:  1,
		length: 0,
		tail:   nil,
		cmp:    cmp,
		index:  make(map[uint64]V),
	}
	var null V
	sl.head = NewNode[V](maxLevel, null)
	for i := 0; i < int(maxLevel); i++ {
		sl.head.SetNext(i, nil)
		sl.head.SetSpan(i, 0)
	}
	sl.head.backward = nil
	return sl
}

func (l *SkipList[V]) Level() uint32 { return l.level }

func (l *SkipList[V]) Length() int { return l.length }

func (l *SkipList[V]) Head() *Node[V] { return l.head }

func (l *SkipList[V]) Tail() *Node[V] { return l.tail }

func (l *SkipList[V]) First() *Node[V] { return l.head.Next(0) }

func (l *SkipList[V]) randomLevel() uint32 {
	level := uint32(1)
	for (rand.Uint32()&0xFFFF) < uint32(P*0xFFFF) && level < maxLevel {
		level++
	}
	return level
}

func (l *SkipList[V]) Insert(value V) (*Node[V], error) {
	if _, ok := l.index[value.Key()]; ok {
		return nil, RepeatedKey
	}
	var update [maxLevel]*Node[V]
	var rank [maxLevel]uint32
	x := l.head
	for i := int(l.level - 1); i >= 0; i-- {
		if i == int(l.level-1) {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for next := x.Next(i); next != nil &&
			(l.cmp.CmpScore(next.value, value) < 0 ||
				(l.cmp.CmpScore(next.value, value) == 0 &&
					l.cmp.CmpKey(next.value, value) < 0)); next = x.Next(i) {
			rank[i] += x.Span(i)
			x = next
		}
		update[i] = x
	}

	level := l.randomLevel()

	if level > l.level {
		for i := l.level; i < level; i++ {
			rank[i] = 0
			update[i] = l.head
			update[i].SetSpan(int(i), uint32(l.length))
		}
		l.level = level
	}

	x = NewNode(level, value)
	for i := 0; i < int(level); i++ {
		x.SetNext(i, update[i].Next(i))
		update[i].SetNext(i, x)

		x.SetSpan(i, update[i].Span(i)-(rank[0]-rank[i]))
		update[i].SetSpan(i, rank[0]-rank[i]+1)
	}

	for i := level; i < l.level; i++ {
		update[i].SetSpan(int(i), update[i].Span(int(i))+1)
	}

	if update[0] == l.head {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.Next(0) != nil {
		x.Next(0).backward = x
	} else {
		l.tail = x
	}
	l.length++
	l.index[value.Key()] = value
	return x, nil
}

func (l *SkipList[V]) deleteNode(x *Node[V], update []*Node[V]) {
	for i := 0; i < int(l.level); i++ {
		if update[i].Next(i) == x {
			update[i].SetSpan(i, update[i].Span(i)+x.Span(i)-1)
			update[i].SetNext(i, x.Next(i))
		} else {
			update[i].SetSpan(i, update[i].Span(i)-1)
		}
	}

	if x.Next(0) != nil {
		x.Next(0).backward = x.backward
	} else {
		l.tail = x.backward
	}

	for l.level > 1 && l.head.Next(int(l.level-1)) == nil {
		l.level--
	}
	l.length--
}

func (l *SkipList[V]) Delete(key uint64) int {
	value, ok := l.index[key]
	if !ok {
		return 0
	}
	update := make([]*Node[V], int(l.level))
	var x = l.head
	for i := int(l.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			(l.cmp.CmpScore(next.value, value) < 0 ||
				(l.cmp.CmpScore(next.value, value) == 0 &&
					l.cmp.CmpKey(next.value, value) < 0)); next = x.Next(i) {
			x = next
		}
		update[i] = x
	}

	x = x.Next(0)
	if x != nil &&
		l.cmp.CmpKey(x.value, value) == 0 &&
		l.cmp.CmpScore(x.value, value) == 0 {
		l.deleteNode(x, update)
		delete(l.index, value.Key())
		return 1
	}
	return 0
}

// GetRank TODO: 1-based rank
func (l *SkipList[V]) GetRank(value V) uint32 {
	var rank uint32 = 0
	x := l.head
	for i := int(l.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			(l.cmp.CmpScore(next.value, value) < 0 ||
				(l.cmp.CmpScore(next.value, value) == 0 &&
					l.cmp.CmpKey(next.value, value) <= 0)); next = x.Next(i) {
			rank += x.Span(i)
			x = next
		}
		if x != l.head && l.cmp.CmpKey(x.value, value) == 0 {
			return rank
		}
	}
	return 0
}

func (l *SkipList[V]) GetNodeByRank(rank uint32) *Node[V] {
	x := l.head
	var traversed uint32 = 0
	for i := int(l.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			traversed+x.Span(i) <= rank; next = x.Next(i) {
			traversed += x.Span(i)
			x = next
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

func (l *SkipList[V]) Range(fun func(v V) bool) {
	for e := l.head.Next(0); e != nil; e = e.Next(0) {
		if !fun(e.Value()) {
			return
		}
	}
}

func (l *SkipList[V]) Dump() {
	fmt.Println("*************SKIP LIST DUMP START*************")
	for i := int(l.level - 1); i >= 0; i-- {
		fmt.Printf("level:--------%v--------\n", i)
		x := l.head
		for x != nil {
			if x == l.head {
				fmt.Printf("Head span: %v\n", x.Span(i))
			} else {
				fmt.Printf("span: %v value : %v\n", x.Span(i), x.Value())
			}
			x = x.Next(i)
		}
	}
	fmt.Println("*************SKIP LIST DUMP END*************")
}

func (l *SkipList[V]) DumpString() string {
	var buffer bytes.Buffer
	for i := int(l.level - 1); i >= 0; i-- {
		buffer.WriteString(fmt.Sprintf("level:--------%v--------\n", i))
		x := l.head
		for x != nil {
			if x == l.head {
				buffer.WriteString(fmt.Sprintf("Head span: %v\n", x.Span(i)))
			} else {
				buffer.WriteString(fmt.Sprintf("span: %v value : %+v\n", x.Span(i), x.Value()))
			}
			x = x.Next(i)
		}
	}
	return buffer.String()
}
