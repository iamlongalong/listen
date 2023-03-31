package listen

import "context"

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
