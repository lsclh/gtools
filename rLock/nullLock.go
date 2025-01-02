package rLock

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
