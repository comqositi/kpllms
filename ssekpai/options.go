package ssekpai

type Option func(*options)

type options struct {
	ctxDoneFunc   func(done any)
	timeOutFunc   func()
	timeOutSecond int
}

func WithCtxDoneFunc(doneFunc func(done any)) Option {
	return func(o *options) {
		o.ctxDoneFunc = doneFunc
	}
}

func WithTimeOutFunc(timeOutFunc func()) Option {
	return func(o *options) {
		o.timeOutFunc = timeOutFunc
	}
}

func WithTimeOutSecond(timeOunt int) Option {
	return func(o *options) {
		o.timeOutSecond = timeOunt
	}
}
