package resizer

import "io"

// countingWriter wraps a writer to count bytes written.
type counterWriter struct {
	writer       io.Writer
	bytesWritten int64
}

func (cw *counterWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	cw.bytesWritten += int64(n)

	return n, err
}
