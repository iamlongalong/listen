package listen

import "context"

type EventListener interface {
	Emit(ctx context.Context, e Event)
}

type Listenable interface {
	Listen(EventListener)
	UnListen(EventListener)
}

type FuncListener func(ctx context.Context, e Event)

func (f FuncListener) Emit(ctx context.Context, e Event) {
	f(ctx, e)
}

type ListenerManager interface {
	EventListener
	Listenable
}
