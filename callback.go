package listen

import "context"

type FuncOnGet = func(ctx context.Context, key string) (v interface{}, err error, ok bool)
type FuncOnSet = func(ctx context.Context, key string, v interface{}) (ver int, err error, ok bool)
type FuncOnDel = func(ctx context.Context, key string) (ver int, err error, ok bool)
type FuncOnRange = func(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool) (nf func(ctx context.Context, key string, v interface{}) bool, ok bool)

type FuncAfterGet = func(ctx context.Context, key string, v interface{}, err error)
type FuncAfterSet = func(ctx context.Context, key string, v interface{}, ver int, err error)
type FuncAfterDel = func(ctx context.Context, key string, ver int, err error)
type FuncAfterRange = func(ctx context.Context, rangedKeys map[string]struct{})

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

type MapCbFunc struct {
	onGet   []FuncOnGet
	onSet   []FuncOnSet
	onDel   []FuncOnDel
	onRange []FuncOnRange

	afterGet   []FuncAfterGet
	afterSet   []FuncAfterSet
	afterDel   []FuncAfterDel
	afterRange []FuncAfterRange
}

func (efh *MapCbFunc) OnGet(ctx context.Context, key string) (v interface{}, err error, ok bool) {
	for _, f := range efh.onGet {
		// 后面的会覆盖前面的
		v, err, ok = f(ctx, key)
	}

	return
}

func (efh *MapCbFunc) OnSet(ctx context.Context, key string, v interface{}) (ver int, err error, ok bool) {
	for _, f := range efh.onSet {
		// 后面的会覆盖前面的
		ver, err, ok = f(ctx, key, v)
	}

	return
}

func (efh *MapCbFunc) OnDel(ctx context.Context, key string) (ver int, err error, ok bool) {
	for _, f := range efh.onDel {
		// 后面的会覆盖前面的
		ver, err, ok = f(ctx, key)
	}

	return
}

func (efh *MapCbFunc) OnRange(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool) (nf func(ctx context.Context, key string, v interface{}) bool, ok bool) {
	for _, of := range efh.onRange {
		// 后面的会覆盖前面的
		nf, ok = of(ctx, f)
	}

	return
}

func (efh *MapCbFunc) AfterGet(ctx context.Context, key string, v interface{}, err error) {
	for _, f := range efh.afterGet {
		f(ctx, key, v, err)
	}
}
func (efh *MapCbFunc) AfterSet(ctx context.Context, key string, v interface{}, ver int, err error) {
	for _, f := range efh.afterSet {
		f(ctx, key, v, ver, err)
	}
}
func (efh *MapCbFunc) AfterDel(ctx context.Context, key string, ver int, err error) {
	for _, f := range efh.afterDel {
		f(ctx, key, ver, err)
	}
}

func (efh *MapCbFunc) AfterRange(ctx context.Context, rangedKeys map[string]struct{}) {
	for _, f := range efh.afterRange {
		f(ctx, rangedKeys)
	}
}

func (efh *MapCbFunc) On(opt *OnOption) {
	if opt.OnGet != nil {
		_, ok := HasMember(opt.OnGet, efh.onGet)
		if ok {
			return
		}

		efh.onGet = append(efh.onGet, opt.OnGet)
	}

	if opt.OnGet != nil {
		_, ok := HasMember(opt.OnGet, efh.onGet)
		if ok {
			return
		}

		efh.onGet = append(efh.onGet, opt.OnGet)
	}

	if opt.OnDel != nil {
		_, ok := HasMember(opt.OnDel, efh.onDel)
		if ok {
			return
		}

		efh.onDel = append(efh.onDel, opt.OnDel)
	}

	if opt.OnRange != nil {
		_, ok := HasMember(opt.OnRange, efh.onRange)
		if ok {
			return
		}

		efh.onRange = append(efh.onRange, opt.OnRange)
	}

	if opt.AfterGet != nil {
		_, ok := HasMember(opt.AfterGet, efh.afterGet)
		if ok {
			return
		}

		efh.afterGet = append(efh.afterGet, opt.AfterGet)
	}

	if opt.AfterSet != nil {
		_, ok := HasMember(opt.AfterSet, efh.afterSet)
		if ok {
			return
		}

		efh.afterSet = append(efh.afterSet, opt.AfterSet)
	}

	if opt.AfterDel != nil {
		_, ok := HasMember(opt.AfterDel, efh.afterDel)
		if ok {
			return
		}

		efh.afterDel = append(efh.afterDel, opt.AfterDel)
	}

	if opt.AfterRange != nil {
		_, ok := HasMember(opt.AfterRange, efh.afterRange)
		if ok {
			return
		}

		efh.afterRange = append(efh.afterRange, opt.AfterRange)
	}
}

func (efh *MapCbFunc) Off(opt *OnOption) {
	if opt.OnGet != nil {
		idx, ok := HasMember(opt.OnGet, efh.onGet)
		if ok {
			// 姑且先用替代的方式
			efh.onGet[idx] = func(ctx context.Context, key string) (v interface{}, err error, ok bool) { return nil, nil, false }
		}
	}

	if opt.OnSet != nil {
		idx, ok := HasMember(opt.OnSet, efh.onSet)
		if ok {
			// 姑且先用替代的方式
			efh.onSet[idx] = func(ctx context.Context, key string, v interface{}) (ver int, err error, ok bool) {
				return 0, nil, false
			}
		}
	}

	if opt.OnDel != nil {
		idx, ok := HasMember(opt.OnDel, efh.onDel)
		if ok {
			// 姑且先用替代的方式
			efh.onDel[idx] = func(ctx context.Context, key string) (ver int, err error, ok bool) {
				return 0, nil, false
			}
		}
	}

	if opt.OnRange != nil {
		idx, ok := HasMember(opt.OnRange, efh.onRange)
		if ok {
			// 姑且先用替代的方式
			efh.onRange[idx] = func(ctx context.Context, f func(ctx context.Context, key string, v interface{}) bool) (nf func(ctx context.Context, key string, v interface{}) bool, ok bool) {
				return func(ctx context.Context, key string, v interface{}) bool { return false }, false
			}
		}
	}

	if opt.AfterGet != nil {
		idx, ok := HasMember(opt.AfterGet, efh.afterGet)
		if ok {
			efh.afterGet[idx] = func(ctx context.Context, key string, v interface{}, err error) {}
		}
	}

	if opt.AfterSet != nil {
		idx, ok := HasMember(opt.AfterSet, efh.afterSet)
		if ok {
			efh.afterSet[idx] = func(ctx context.Context, key string, v interface{}, ver int, err error) {}
		}
	}

	if opt.AfterDel != nil {
		idx, ok := HasMember(opt.AfterDel, efh.afterDel)
		if ok {
			efh.afterDel[idx] = func(ctx context.Context, key string, ver int, err error) {}
		}
	}

	if opt.AfterRange != nil {
		idx, ok := HasMember(opt.AfterRange, efh.afterRange)
		if ok {
			efh.afterRange[idx] = func(ctx context.Context, rangedKeys map[string]struct{}) {}
		}
	}
}
