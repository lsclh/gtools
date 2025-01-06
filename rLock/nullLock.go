package rLock

type nullLock struct{}

func (r *nullLock) TryLock() bool {
	return true
}

func (r *nullLock) Lock() {
	return
}

func (r *nullLock) Unlock() bool {
	return true
}
