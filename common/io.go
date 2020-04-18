package common

import (
	"io"
)

func ReaderSize(reader io.Reader) (int64, error) {
	var size int64 = 0

	buffer := make([]byte, 1024)
	for {
		if count, err := reader.Read(buffer); err == nil {
			size += int64(count)
		} else if err == io.EOF {
			break
		} else {
			return 0, err
		}
	}

	return size, nil
}
