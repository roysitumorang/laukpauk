package pools

import (
	"bytes"
	"strings"
	"sync"
)

var BufferPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

var BuilderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}
