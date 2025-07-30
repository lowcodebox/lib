package utils_test

import (
	"bytes"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSeekWrapper(t *testing.T) {
	// 1. Создаём базовый io.ReadSeeker
	data := []byte("example content")
	reader := bytes.NewReader(data)

	// 2. Оборачиваем его в ReadSeekWrapper с io.NopCloser
	wrapper := &utils.ReadSeekWrapper{
		ReadSeeker: reader,
		Closer:     io.NopCloser(nil),
	}

	// 3. Проверяем чтение
	buf := make([]byte, 7)
	n, err := wrapper.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, []byte("example"), buf)

	// 4. Проверяем Seek
	offset, err := wrapper.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)

	// 5. Проверяем чтение после Seek
	buf2 := make([]byte, len(data))
	n2, err := wrapper.Read(buf2)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n2)
	assert.Equal(t, data, buf2)

	// 6. Проверяем Close
	err = wrapper.Close()
	assert.NoError(t, err)
}
