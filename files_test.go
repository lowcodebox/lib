package lib_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	lib "git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "foo.txt")

	// Файл ещё не существует
	assert.NoError(t, lib.CreateFile(f))
	// Теперь должен быть пустой файл
	info, err := os.Stat(f)
	assert.NoError(t, err)
	assert.Zero(t, info.Size())

	// Напишем в файл и заново вызовем CreateFile — старый файл должен быть затиран
	assert.NoError(t, ioutil.WriteFile(f, []byte("data"), 0644))
	assert.NoError(t, lib.CreateFile(f))
	info2, err := os.Stat(f)
	assert.NoError(t, err)
	assert.Zero(t, info2.Size())
}

func TestWriteAndReadFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "bar.txt")
	data := []byte("hello world")

	// WriteFile создает и пишет
	assert.NoError(t, lib.WriteFile(f, data))
	// ReadFile читает обратно
	s, err := lib.ReadFile(f)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", s)
}

func TestReadFile_NotExist(t *testing.T) {
	t.Parallel()
	_, err := lib.ReadFile("no_such_file.xyz")
	assert.Error(t, err)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := []byte("foobar")

	assert.NoError(t, ioutil.WriteFile(src, content, 0644))
	assert.NoError(t, lib.CopyFile(src, dst))

	got, err := ioutil.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestCopyFolder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// создаём структуру
	src := filepath.Join(root, "src")
	sub := filepath.Join(src, "subdir")
	assert.NoError(t, os.MkdirAll(sub, 0755))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(src, "a.txt"), []byte("A"), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(sub, "b.txt"), []byte("B"), 0644))

	dst := filepath.Join(root, "dst")
	assert.NoError(t, lib.CopyFolder(src, dst))

	// Проверяем оба файла
	a, err := ioutil.ReadFile(filepath.Join(dst, "a.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "A", string(a))

	b, err := ioutil.ReadFile(filepath.Join(dst, "subdir", "b.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "B", string(b))
}

func TestIsExist(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "f")
	assert.False(t, lib.IsExist(f))

	assert.NoError(t, ioutil.WriteFile(f, []byte{}, 0644))
	assert.True(t, lib.IsExist(f))

	d := filepath.Join(dir, "d")
	assert.NoError(t, os.Mkdir(d, 0755))
	assert.True(t, lib.IsExist(d))
}

func TestCreateDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	nested := filepath.Join(dir, "one", "two", "three")
	assert.NoError(t, lib.CreateDir(nested, 0750))
	info, err := os.Stat(nested)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	// Права могут быть урезаны umask-ом, но папка должна существовать
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "todel.txt")
	assert.NoError(t, ioutil.WriteFile(f, []byte("X"), 0644))
	assert.True(t, lib.IsExist(f))

	assert.NoError(t, lib.DeleteFile(f))
	assert.False(t, lib.IsExist(f))
}

func TestMoveFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "orig.txt")
	dst := filepath.Join(dir, "moved.txt")
	content := []byte("MOVE!")

	assert.NoError(t, ioutil.WriteFile(src, content, 0644))
	assert.NoError(t, lib.MoveFile(src, dst))

	// Исходного файла нет, новый есть с тем же содержимым
	assert.False(t, lib.IsExist(src))
	got, err := ioutil.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestZipAndUnzip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	src := filepath.Join(root, "zipsrc")
	sub := filepath.Join(src, "sub")
	assert.NoError(t, os.MkdirAll(sub, 0755))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(src, "f1.txt"), []byte("F1"), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(sub, "f2.txt"), []byte("F2"), 0644))

	zipFile := filepath.Join(root, "out.zip")
	assert.NoError(t, lib.Zip(src, zipFile))
	assert.True(t, lib.IsExist(zipFile))

	unz := filepath.Join(root, "unz")
	assert.NoError(t, lib.Unzip(zipFile, unz))
	// Проверяем файлы после распаковки
	got1, err := ioutil.ReadFile(filepath.Join(unz, filepath.Base(src), "f1.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "F1", string(got1))

	got2, err := ioutil.ReadFile(filepath.Join(unz, filepath.Base(src), "sub", "f2.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "F2", string(got2))
}

func TestChmod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "perm.txt")
	assert.NoError(t, ioutil.WriteFile(f, []byte("x"), 0644))

	// Ставим «исполняемый» бит
	assert.NoError(t, lib.Chmod(f, 0755))
	info, err := os.Stat(f)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}
