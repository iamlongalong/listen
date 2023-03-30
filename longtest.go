package listen

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrNotExsits = errors.New("value not exsits")
)

type IMap interface {
	Get(ctx context.Context, key string) (v interface{}, err error)
	Set(ctx context.Context, key string, v interface{}) (ver int, err error)
	Del(ctx context.Context, key string) (ver int, err error)
	Range(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool)

	On(opt *OnOption)
	Off(opt *OnOption)

	Listen(EventListener)
	UnListen(EventListener)

	Mashal(ctx context.Context) ([]byte, error)
	UnMashal(ctx context.Context, b []byte) error

	GetVersion(ctx context.Context) (int, error)
	GetEvents(ctx context.Context, opt getLogOption) ([]Event, error)

	Sync(ctx context.Context, es []Event) error
}

type FuncOnGet = func(ctx context.Context, key string) (v interface{}, err error, ok bool)
type FuncOnSet = func(ctx context.Context, key string, v interface{}) (ver int, err error, ok bool)
type FuncOnDel = func(ctx context.Context, key string) (ver int, err error, ok bool)
type FuncOnRange = func(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool) (nf func(ctx context.Context, key string, v interface{}) bool, ok bool)

type FuncAfterGet = func(ctx context.Context, key string, v interface{}, err error)
type FuncAfterSet = func(ctx context.Context, key string, v interface{}, ver int, err error)
type FuncAfterDel = func(ctx context.Context, key string, ver int, err error)
type FuncAfterRange = func(ctx context.Context, rangedKeys map[string]struct{})

type FuncListener func(ctx context.Context, e Event)

func (f FuncListener) Emit(ctx context.Context, e Event) {
	f(ctx, e)
}

type OnOption struct {
	OnGet   FuncOnGet
	OnSet   FuncOnSet
	OnDel   FuncOnDel
	OnRange FuncOnRange

	AfterGet   FuncAfterGet
	AfterSet   FuncAfterSet
	AfterDel   FuncAfterDel
	AfterRange FuncAfterRange
}

type EventOp int

const (
	EventMapSet EventOp = iota + 1
	EventMapDel
)

type Event struct {
	Option  EventOp
	Version int

	Payload Encoder
}

type IEventManager interface {
	IEventStorage
	IEvnetGetter
}

type IEventStorage interface {
	AppendEvent(ctx context.Context, e Event) (ver int, err error)
}

type IEvnetGetter interface {
	GetEvents(ctx context.Context, opt getLogOption) (es []Event, err error)
	Version(ctx context.Context) (ver int, err error)
}

type EventListener interface {
	Emit(ctx context.Context, e Event)
}

type ListenerHub interface {
	Listen(EventListener)
	UnListen(EventListener)
}

type ListenerManager interface {
	EventListener
	ListenerHub
}

// MapEventHub implement EventListener for listener management
type MapEventHub struct {
	evGetter IEvnetGetter

	mu sync.Mutex

	listeners map[EventListener]struct{}
}

func (meh *MapEventHub) Emit(ctx context.Context, e Event) {
	meh.mu.Unlock()
	defer meh.mu.Unlock()

	for l := range meh.listeners {
		// TODO 最好处理顺序问题
		l.Emit(ctx, e)
	}
}

func (meh *MapEventHub) ListenAndSync(el EventListener, after int) {
	defer meh.Listen(el)

	ctx := context.Background()

	es, err := meh.evGetter.GetEvents(ctx, LogOptionGreaterThan(after, true))
	if err != nil {
		// TODO handle err
		return
	}

	for _, e := range es {
		el.Emit(ctx, e)
	}
}

func (meh *MapEventHub) Listen(el EventListener) {
	meh.mu.Lock()
	defer meh.mu.Unlock()

	meh.listeners[el] = struct{}{}
}

func (meh *MapEventHub) UnListen(el EventListener) {
	meh.mu.Lock()
	defer meh.mu.Unlock()

	delete(meh.listeners, el)
}

func NewXMap() IMap {
	return &XMap{
		m: map[string]interface{}{},

		ListenerManager: &MapEventHub{},
		MapCbFunc:       MapCbFunc{},
		logs:            &MapEventManager{},
	}
}

type XMap struct {
	mu sync.Mutex

	m map[string]interface{}

	MapCbFunc

	ListenerManager

	logs IEventManager
}

func (xm *XMap) Get(ctx context.Context, key string) (v interface{}, err error) {
	var ok bool
	defer func() {
		xm.AfterGet(ctx, key, v, err)
	}()

	// 拦截
	v, err, ok = xm.MapCbFunc.OnGet(ctx, key)
	if ok {
		return v, err
	}

	v, ok = xm.m[key]
	if ok {
		return v, nil
	}

	return nil, ErrNotExsits
}

func (xm *XMap) Set(ctx context.Context, key string, v interface{}) (ver int, err error) {
	var ok bool
	defer func() {
		xm.AfterSet(ctx, key, v, ver, err)
	}()

	// 拦截
	ver, err, ok = xm.MapCbFunc.OnSet(ctx, key, v)
	if ok {
		return ver, err
	}

	e := Event{
		Option:  EventMapSet,
		Payload: nil, // TODO
	}

	ver, err = xm.logs.AppendEvent(ctx, e)

	e.Version = ver

	// 发送事件
	defer xm.ListenerManager.Emit(ctx, e)

	xm.set(ctx, key, v)

	return ver, err
}

func (xm *XMap) set(ctx context.Context, key string, v interface{}) {
	xm.m[key] = v
}

func (xm *XMap) Del(ctx context.Context, key string) (ver int, err error) {
	var ok bool
	defer func() {
		xm.AfterDel(ctx, key, ver, err)
	}()

	// 拦截
	ver, err, ok = xm.MapCbFunc.OnDel(ctx, key)
	if ok {
		return ver, err
	}

	e := Event{
		Option:  EventMapDel,
		Payload: nil, // TODO
	}

	ver, err = xm.logs.AppendEvent(ctx, e)

	e.Version = ver

	// 发送事件
	defer xm.ListenerManager.Emit(ctx, e)

	xm.del(ctx, key)

	return ver, err

}

func (xm *XMap) del(ctx context.Context, key string) {
	delete(xm.m, key)
}

func (xm *XMap) Range(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool) {
	var ok bool
	rangedKeys := make(map[string]struct{}, 0)

	defer func() {
		xm.AfterRange(ctx, rangedKeys)
	}()

	var xf func(ctx context.Context, key string, v interface{}) bool
	xf = f

	// 拦截
	nf, ok := xm.MapCbFunc.OnRange(ctx, f)
	if ok {
		xf = nf
	}

	for k, v := range xm.m {
		rangedKeys[k] = struct{}{}

		finish := xf(ctx, k, v)
		if finish {
			return
		}
	}
}

func (xm *XMap) Mashal(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (xm *XMap) UnMashal(ctx context.Context, b []byte) error {
	return nil
}

func (xm *XMap) GetVersion(ctx context.Context) (int, error) {
	return xm.logs.Version(ctx)
}

func (xm *XMap) GetEvents(ctx context.Context, opt getLogOption) ([]Event, error) {
	return xm.logs.GetEvents(ctx, opt)
}

func (xm *XMap) Sync(ctx context.Context, es []Event) error {
	// TODO
	return nil
}
