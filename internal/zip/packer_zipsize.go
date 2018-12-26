package zip

import "io"

func ZipSize(r io.ReaderAt, size int64) (int64, error) {
	dir, err := readDirectoryEnd(r, size)
	if err != nil {
		return 0, err
	}

	zipsize := int64(dir.directoryOffset) + int64(dir.directorySize) + 22 + int64(dir.commentLen)
	return zipsize, nil
}
