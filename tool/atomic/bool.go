package atomic

import (
	"github.com/Dongxiem/gfaio/tool/atomic"
)

type Bool struct {
	b int32
}

// Set：将 a.b 置位为 1
func (a *Bool) Set(b bool) bool {
	var newV int32
	if b {
		newV = 1
	}
	return atomic.SwapInt32(&a.b, newV) == 1
}

// Get：原子判断 a.b 是否为 1
func (a *Bool) Get() bool {
	return atomic.LoadInt32(&a.b) == 1
}
