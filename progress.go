package request

import "io"

// ProgressBarMaxSetter is an interface that allows setting the maximum value of a progress bar
type ProgressBarMaxSetter interface {
	SetMax64(int64)
}

// ProgressBarMaxChanger is an interface that allows setting the maximum value of a progress bar
//
// This interface allows packages such as "/github.com/schollz/progressbar/v3" to be used as progress bars
type ProgressBarMaxChanger interface {
	ChangeMax64(int64)
}

type progressReader struct {
	io.Reader
	Progress io.Writer
}

func (reader *progressReader) Read(p []byte) (n int, err error) {
	n, err = reader.Reader.Read(p)
	_, _ = reader.Progress.Write(p[:n])
	return
}
