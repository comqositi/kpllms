package ssekpai

type Option func(*options)

type options struct {
	ctxDoneFunc func(done any)
	timeOutFunc func()
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
