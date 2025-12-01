package cronlogger

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func ReadStdin() (string, error) {
	finfo, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if finfo.Mode()&os.ModeNamedPipe == 0 {
		return "", fmt.Errorf("stdin must be a pipe")
	}

	r := bufio.NewReader(os.Stdin)
	readBuf := make([]byte, 4*1024)
	var result []byte

	for {
		n, err := r.Read(readBuf)
		if n > 0 {
			result = append(result, readBuf[:n]...)
		}

		if err != nil {
			if err == io.EOF {
				return string(result), nil
			}
			return "", fmt.Errorf("cannot read input; %v", err)
		}
	}
}
