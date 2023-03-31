package listen

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrNotExsits = errors.New("value not exsits")
)

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
