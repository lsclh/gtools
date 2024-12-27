package rLock

func NewNullLock() *NullLock {
	return &NullLock{}
}

type NullLock struct{}

func (r *NullLock) TryLock() bool {
	return true
}

func (r *NullLock) Lock() {
	return
}

func (r *NullLock) Unlock() bool {
	return true
}
