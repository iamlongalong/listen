package listen

import (
	"context"
	"sync"
)

type getLogOption struct {
	operation string // == > >= < <=  ~ (latest n)  <=x<  <x<=  <x<  <=x<=
	val1      int

	val2 int
}

func LogOptionEqualVersion(v int) getLogOption {
	return getLogOption{operation: "==", val1: v}
}

func LogOptionGreaterThan(v int, withEqual bool) getLogOption {
	if withEqual {
		return getLogOption{operation: ">=", val1: v}
	}
	return getLogOption{operation: ">", val1: v}
}

func LogOptionLessThan(v int, withEqual bool) getLogOption {
	if withEqual {
		return getLogOption{operation: "<=", val1: v}
	}
	return getLogOption{operation: "<", val1: v}
}

func LogOptionBetween(greaterThan int, equalGreater bool, lessThan int, equalLess bool) getLogOption {
	opt := getLogOption{
		val1: lessThan,
		val2: greaterThan,
	}

	if equalLess {
		if equalGreater {
			opt.operation = "<=x<="
			return opt
		} else {
			opt.operation = "<x<="
			return opt
		}
	} else {
		if equalGreater {
			opt.operation = "<=x<"
			return opt
		} else {
			opt.operation = "<x<"
			return opt
		}
	}
}

func (opt *getLogOption) Equal() (ok bool, val int) {
	if opt.operation == "==" {
		return true, opt.val1
	}

	return false, 0
}

func (opt *getLogOption) Latest() (ok bool, val int) {
	if opt.operation == "~" {
		return true, opt.val1
	}

	return false, 0
}

func (opt *getLogOption) GreaterThan() (ok bool, gtval int, withEqual bool) {
	i, isin := IsStrInList(opt.operation, []string{">", ">=", "<=x<=", "<x<=", "<=x<", "<x<"})
	if isin {
		return true, opt.val1, i == 1 || i == 2 || i == 4
	}

	return false, 0, false
}

func (opt *getLogOption) LessThan() (ok bool, ltval int, withEqual bool) {
	i, isin := IsStrInList(opt.operation, []string{"<", "<=", "<=x<=", "<x<=", "<=x<", "<x<"})

	if isin {
		withEqual = (i == 1 || i == 2 || i == 3)
		if i > 1 {
			return true, opt.val2, withEqual
		}

		return true, opt.val1, withEqual
	}

	return false, 0, false
}

// MapEventManager in mem log
type MapEventManager struct {
	mu sync.Mutex

	logs []Event

	version int
}

func (mm *MapEventManager) AppendEvent(ctx context.Context, e Event) (ver int, err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.logs = append(mm.logs, e)
	mm.version++

	return mm.version, nil
}

func (mm *MapEventManager) GetEvents(ctx context.Context, opt getLogOption) (es []Event, err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// 精准匹配
	ok, v := opt.Equal()
	if ok {
		if v > mm.version || v < 0 {
			return []Event{}, nil
		}

		return []Event{mm.logs[opt.val1]}, nil
	}

	// 最近多个
	ok, v = opt.Latest()
	if ok {
		es = make([]Event, opt.val1)
		copy(es, mm.logs[mm.version-opt.val1:])
		return es, nil
	}

	// 范围
	lessok, lessVal, lessEqual := opt.LessThan()
	greaterok, greaterVal, greaterEqual := opt.GreaterThan()

	if lessok {
		if v < 0 { // 小于0过滤
			return []Event{}, nil
		}

		if lessVal > mm.version {
			lessVal = mm.version // 最大仅为 version 的量
		}
	} else {
		lessVal = mm.version
		lessEqual = true
	}

	if greaterok {
		if greaterVal-1 > mm.version {
			return []Event{}, nil // 大于最大 version 过滤
		}
		if greaterVal < 0 {
			greaterVal = 0 // 最小为0
		}
	} else {
		greaterVal = mm.version
		greaterEqual = true
	}

	eqNums := 0
	if lessEqual {
		eqNums++
	}
	if greaterEqual {
		eqNums++
	}

	es = make([]Event, greaterVal-lessVal+eqNums)
	copy(es, mm.logs[lessVal:greaterVal]) // TODO 调试边界情况

	return es, nil
}

func (mm *MapEventManager) Version(ctx context.Context) (ver int, err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	return mm.version, nil
}
