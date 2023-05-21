package lib

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CreateFile Создаем файл по указанному пути если его нет
func CreateFile(path string) (err error) {

	// detect if file exists
	_, err = os.Stat(path)
	var file *os.File

	// delete old file if exists
	if !os.IsNotExist(err) {
		os.RemoveAll(path)
	}

	// create file
	file, err = os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return err
}

// WriteFile пишем в файл по указанному пути
func WriteFile(path string, data []byte) (err error) {

	// detect if file exists and create
	err = CreateFile(path)
	if err != nil {
		return
	}

	// open file using READ & WRITE permission
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	// write into file
	_, err = file.Write(data)
	if err != nil {
		return
	}

	// save changes
	err = file.Sync()
	if err != nil {
		return
	}

	return
}

// ReadFile читаем файл. (отключил: всегда в рамках рабочей диретории)
func ReadFile(path string) (result string, err error) {
	// если не от корня, то подставляем текущую директорию
	//if path[:1] != "/" {
	//	path = CurrentDir() + "/" + path
	//} else {
	//	path = CurrentDir() + path
	//}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err == nil {
		result = string(b)
	}

	return result, err
}

// CopyFolder копирование папки
func CopyFolder(source string, dest string) (err error) {

	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	defer directory.Close()
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			err = CopyFolder(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

// CopyFile копирование файла
func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return
}

// IsExist определяем наличие директории/файла
func IsExist(path string) (exist bool) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}

	return false
}

// CreateDir создание папки
func CreateDir(path string, mode os.FileMode) (err error) {
	if mode == 0 {
		mode = 0711
	}
	err = os.MkdirAll(path, mode)
	if err != nil {
		return err
	}

	return nil
}

func DeleteFile(path string) (err error) {
	err = os.Remove(path)
	if err != nil {
		return
	}

	return nil
}

func MoveFile(source string, dest string) (err error) {
	err = CopyFile(source, dest)
	if err != nil {
		return
	}
	err = DeleteFile(source)
	if err != nil {
		return
	}

	return nil
}

// Zip
// zip("/tmp/documents", "/tmp/backup.zip")
func Zip(source, target string) (err error) {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// Unzip
// unzip("/tmp/report-2015.zip", "/tmp/reports/")
func Unzip(archive, target string) (err error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}
	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return
}

func Chmod(path string, mode os.FileMode) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Chmod(mode)

	return err
}
