package skiplist

import (
	"fmt"
	"math/rand"
)

var SKIPLIST_P float32 = 0.25

const SKIPLIST_MAXLEVEL int = 32

type Comparatorer interface {
	CmpScore(interface{}, interface{}) int
	CmpKey(interface{}, interface{}) int
}

type SkipListLevel struct {
	forward *SkipListNode
	span    uint32
}

type SkipListNode struct {
	value    interface{}
	backward *SkipListNode
	level    []SkipListLevel
}

func NewSkipListNode(level int, value interface{}) *SkipListNode {
	sln := &SkipListNode{
		value: value,
		level: make([]SkipListLevel, level),
	}
	return sln
}

func (this *SkipListNode) Next(i int) *SkipListNode {
	return this.level[i].forward
}

func (this *SkipListNode) SetNext(i int, next *SkipListNode) {
	this.level[i].forward = next
}

func (this *SkipListNode) Span(i int) uint32 {
	return this.level[i].span
}

func (this *SkipListNode) SetSpan(i int, span uint32) {
	this.level[i].span = span
}

func (this *SkipListNode) Value() interface{} {
	return this.value
}

func (this *SkipListNode) Prev() *SkipListNode {
	return this.backward
}

// Valuer /
type Valuer interface {
	Key() uint64
	Score() uint64
	ReCalcScore()
}

// Set
type Set struct {
	head, tail    *SkipListNode
	length, level uint32
	Comparatorer

	index map[uint64]Valuer
}

func NewSet(cmp Comparatorer) *Set {
	sl := &Set{
		level:        1,
		length:       0,
		tail:         nil,
		Comparatorer: cmp,
		index:        make(map[uint64]Valuer),
	}
	sl.head = NewSkipListNode(SKIPLIST_MAXLEVEL, nil)
	for i := 0; i < SKIPLIST_MAXLEVEL; i++ {
		sl.head.SetNext(i, nil)
		sl.head.SetSpan(i, 0)
	}
	sl.head.backward = nil
	return sl
}

func (this *Set) Level() uint32 { return this.level }

func (this *Set) Length() uint32 { return this.length }

func (this *Set) Head() *SkipListNode { return this.head }

func (this *Set) Tail() *SkipListNode { return this.tail }

func (this *Set) First() *SkipListNode { return this.head.Next(0) }

func (this *Set) randomLevel() int {
	level := 1
	for (rand.Uint32()&0xFFFF) < uint32(SKIPLIST_P*0xFFFF) && level < SKIPLIST_MAXLEVEL {
		level++
	}
	return level
}

//Insert (先调用删除,在调用insert)
//必须确保删除和添加时候的key和score是和之前的一致
func (this *Set) Insert(value Valuer) {
	this.Delete(value)
	value.ReCalcScore()

	//add new
	var update [SKIPLIST_MAXLEVEL]*SkipListNode
	var rank [SKIPLIST_MAXLEVEL]uint32
	x := this.head
	for i := int(this.level - 1); i >= 0; i-- {
		if i == int(this.level-1) {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for next := x.Next(i); next != nil &&
			(this.CmpScore(next.value, value) < 0 ||
				(this.CmpScore(next.value, value) == 0 &&
					this.CmpKey(next.value, value) < 0)); next = x.Next(i) {
			rank[i] += x.Span(i)
			x = next
		}
		update[i] = x
	}

	level := uint32(this.randomLevel())

	if level > this.level {
		for i := this.level; i < level; i++ {
			rank[i] = 0
			update[i] = this.head
			update[i].SetSpan(int(i), this.length)
		}
		this.level = level
	}

	x = NewSkipListNode(int(level), value)
	for i := 0; i < int(level); i++ {
		x.SetNext(i, update[i].Next(i))
		update[i].SetNext(i, x)

		x.SetSpan(i, update[i].Span(i)-(rank[0]-rank[i]))
		update[i].SetSpan(i, rank[0]-rank[i]+1)
	}

	for i := level; i < this.level; i++ {
		update[i].SetSpan(int(i), update[i].Span(int(i))+1)
	}

	if update[0] == this.head {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.Next(0) != nil {
		x.Next(0).backward = x
	} else {
		this.tail = x
	}
	this.length++

	this.index[value.Key()] = value
}

func (this *Set) GetElement(key uint64) Valuer {
	if value, exist := this.index[key]; exist {
		return value
	}
	return nil
}

func (this *Set) Delete(value Valuer) int {
	if value, exist := this.index[value.Key()]; exist {
		delete(this.index, value.Key())

		//delete
		update := make([]*SkipListNode, int(this.level))
		var x *SkipListNode = this.head
		for i := int(this.level - 1); i >= 0; i-- {
			for next := x.Next(i); next != nil &&
				(this.CmpScore(next.value, value) < 0 ||
					(this.CmpScore(next.value, value) == 0 &&
						this.CmpKey(next.value, value) < 0)); next = x.Next(i) {
				x = next
			}
			update[i] = x
		}

		x = x.Next(0)
		if x != nil &&
			this.CmpKey(x.value, value) == 0 &&
			this.CmpScore(x.value, value) == 0 {
			this.deleteNode(x, update)
			return 1
		}
		return 0
	}
	return 0
}

func (this *Set) DeleteElement(key uint64) {
	if value, exist := this.index[key]; exist {
		this.Delete(value)
	}
}

func (this *Set) GetByRank(rank uint32) interface{} {
	v := this.GetNodeByRank(rank)
	if v == nil {
		return nil
	}
	return v.Value()
}

func (this *Set) GetRank(key uint64) uint32 {
	if value, exist := this.index[key]; exist {
		var rank uint32 = 0
		x := this.head
		for i := int(this.level - 1); i >= 0; i-- {
			for next := x.Next(i); next != nil &&
				(this.CmpScore(next.value, value) < 0 ||
					(this.CmpScore(next.value, value) == 0 &&
						this.CmpKey(next.value, value) <= 0)); next = x.Next(i) {
				rank += x.Span(i)
				x = next
			}
			if x != this.head && this.CmpKey(x.value, value) == 0 {
				return rank
			}
		}
		return 0
	}
	return 0
}

func (this *Set) GetNodeByRank(rank uint32) *SkipListNode {
	x := this.head
	var traversed uint32 = 0
	for i := int(this.level - 1); i >= 0; i-- {
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

func (this *Set) DeleteRangeByRank(start, end uint32) uint32 {
	level := int(this.Level())
	update := make([]*SkipListNode, level)
	var removed uint32 = 0
	var traversed uint32 = 0
	x := this.Head()
	for i := level - 1; i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			x.Span(i)+traversed < start; next = x.Next(i) {
			traversed += x.Span(i)
			x = next
		}
		update[i] = x
	}
	x = x.Next(0)
	traversed++
	for x != nil && traversed <= end {
		next := x.Next(0)
		this.deleteNode(x, update)
		delete(this.index, x.Value().(Valuer).Key())
		removed++
		traversed++
		x = next
	}
	return removed
}

func (this *Set) Dump() {
	fmt.Println("*************SKIP LIST DUMP START*************")
	for i := int(this.level - 1); i >= 0; i-- {
		fmt.Printf("level:--------%v--------\n", i)
		x := this.head
		for x != nil {
			if x == this.head {
				fmt.Printf("Head span: %v\n", x.Span(i))
			} else {
				fmt.Printf("span: %v value : %v\n", x.Span(i), x.Value())
			}
			x = x.Next(i)
		}
	}
	fmt.Println("*************SKIP LIST DUMP END*************")
}

//1-based rank
func (this *Set) GetRightRange(start, end uint32, reversal bool) (uint32, uint32) {
	length := this.Length()
	if length == 0 || start == 0 || end < start || start > length {
		return 0, 0
	}
	if reversal {
		start = length + 1 - start
		if end > length {
			end = 1
		} else {
			end = length + 1 - end
		}
	} else {
		if end > length {
			end = length
		}
	}
	return start, end
}

// GetRange return 1-based elements in [start, end]
func (this *Set) GetRange(start uint32, end uint32, reverse bool) []interface{} {
	// var retKey []uint64
	// var retScore []uint64
	var out []interface{}
	if start == 0 {
		start = 1
	}
	if end == 0 {
		end = this.Length()
	}
	if start > end || start > this.Length() {
		return out
	}
	if end > this.Length() {
		end = this.Length()
	}
	rangeLen := end - start + 1
	if reverse {
		node := this.GetNodeByRank(this.Length() - start + 1)
		for i := uint32(0); i < rangeLen; i++ {
			// retKey = append(retKey, node.Value().(Valuer).Key())
			// retScore = append(retScore, node.Value().(Valuer).Score())
			out = append(out, node.Value())
			node = node.backward
		}
	} else {
		node := this.GetNodeByRank(start)
		for i := uint32(0); i < rangeLen; i++ {
			// retKey = append(retKey, node.Value().(Valuer).Key())
			// retScore = append(retScore, node.Value().(Valuer).Score())
			out = append(out, node.Value())
			node = node.level[0].forward
		}
	}
	// return retKey, retScore
	return out
}

type RangeSpec struct {
	MinEx, MaxEx bool
	Min, Max     uint64
}

func (this *Set) ValueGteMin(value uint64, spec *RangeSpec) bool {
	if spec.MinEx {
		return value > spec.Min
	}
	return value >= spec.Min
}

func (this *Set) ValueLteMax(value uint64, spec *RangeSpec) bool {
	if spec.MaxEx {
		return value < spec.Max
	}
	return value <= spec.Max
}

func (this *Set) IsInRange(rg *RangeSpec) bool {
	if rg.Min > rg.Max ||
		(rg.Min == rg.Max && (rg.MinEx || rg.MaxEx)) {
		return false
	}

	x := this.Tail()
	if x == nil || !this.ValueGteMin(x.Value().(Valuer).Score(), rg) {
		return false
	}

	x = this.First()
	if x == nil || !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
		return false
	}
	return true
}

func (this *Set) FirstInRange(rg *RangeSpec) *SkipListNode {
	if !this.IsInRange(rg) {
		return nil
	}

	x := this.Head()
	for i := int(this.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			!this.ValueGteMin(next.Value().(Valuer).Score(), rg); next = x.Next(i) {
			x = next
		}
	}
	x = x.Next(0)
	if !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
		return nil
	}
	return x
}

func (this *Set) LastInRange(rg *RangeSpec) *SkipListNode {
	if !this.IsInRange(rg) {
		return nil
	}

	x := this.Head()
	for i := int(this.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			this.ValueLteMax(next.Value().(Valuer).Score(), rg); next = x.Next(i) {
			x = next
		}
	}
	if !this.ValueGteMin(x.Value().(Valuer).Score(), rg) {
		return nil
	}
	return x
}

func (this *Set) DeleteRangeByScore(rg *RangeSpec) uint32 {
	update := make([]*SkipListNode, int(this.Level()))
	var removed uint32 = 0
	x := this.Head()
	for i := int(this.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			((rg.MinEx && next.Value().(Valuer).Score() <= rg.Min) ||
				(!rg.MinEx && next.Value().(Valuer).Score() < rg.Min)); next = x.Next(i) {
			x = next
		}
		update[i] = x
	}
	x = x.Next(0)
	for x != nil &&
		((rg.MaxEx && x.Value().(Valuer).Score() < rg.Max) ||
			(!rg.MaxEx && x.Value().(Valuer).Score() <= rg.Max)) {
		next := x.Next(0)
		this.deleteNode(x, update)
		delete(this.index, x.Value().(Valuer).Key())
		removed++
		x = next
	}
	return removed
}

func (this *Set) GetRangeByScore(rg *RangeSpec) []interface{} {
	var values []interface{}
	x := this.FirstInRange(rg)
	for x != nil {
		if !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
			break
		}
		values = append(values, x.value)
		x = x.Next(0)
	}
	return values
}

func (this *Set) Range(f func(interface{})) {
	for tmp := this.First(); tmp != nil; tmp = tmp.Next(0) {
		f(tmp.Value())
	}
}

func (this *Set) deleteNode(x *SkipListNode, update []*SkipListNode) {
	for i := 0; i < int(this.level); i++ {
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
		this.tail = x.backward
	}

	for this.level > 1 && this.head.Next(int(this.level-1)) == nil {
		this.level--
	}
	this.length--
}
