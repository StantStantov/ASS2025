package framebuffer

import (
	"strings"
)

type Buffer struct {
	Buffers []strings.Builder
	First   *strings.Builder
	Second  *strings.Builder
	Third   *strings.Builder
	Fourth  *strings.Builder
}

func InitBuffer(buffer *Buffer) {
	buffer.Buffers = make([]strings.Builder, 4)
	buffer.First = &buffer.Buffers[0]
	buffer.Second = &buffer.Buffers[1]
	buffer.Third = &buffer.Buffers[2]
	buffer.Fourth = &buffer.Buffers[3]
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	return b.Second.Write(p)
}

func String(b *Buffer, setBuffer *strings.Builder) {
	setBuffer.WriteString(b.Third.String())
	setBuffer.WriteString(b.Fourth.String())
	setBuffer.WriteString(b.First.String())
}

func Next(b *Buffer) {
	b.First, b.Second, b.Third, b.Fourth = b.Second, b.Third, b.Fourth, b.First
	b.Second.Reset()
}
