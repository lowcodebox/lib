package lib

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type ConfigFileLogger struct {
	Dir                                string
	IntervalReload, IntervalClearFiles time.Duration
	PeriodSaveFiles                    string
}

// вспомогательная фукнция очистки старых файлов для файлового логера
func (l *log) fileLoggerClearing(ctx context.Context) {

	// попытка очистки старых файлов (каждые пол часа)
	go func() {
		ticker := time.NewTicker(l.IntervalClearFiles)
		defer ticker.Stop()

		// получаем период, через который мы будем удалять файлы
		period := l.PeriodSaveFiles
		if period == "" {
			l.Error(fmt.Errorf("%s", "Fail perion save log files. (expected format: year-month-day; eg: 0-1-0)"))
			return
		}
		slPeriod := strings.Split(period, "-")
		if len(slPeriod) < 3 {
			l.Error(fmt.Errorf("%s", "Fail perion save log files. (expected format: year-month-day; eg: 0-1-0)"))
			return
		}

		// получаем числовые значения года месяца и дня для расчета даты удаления файлов
		year, err := strconv.Atoi(slPeriod[0])
		if err != nil {
			l.Error(err, "Fail converted Year from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}
		month, err := strconv.Atoi(slPeriod[1])
		if err != nil {
			l.Error(err, "Fail converted Month from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}
		day, err := strconv.Atoi(slPeriod[2])
		if err != nil {
			l.Error(err, "Fail converted Day from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				oneMonthAgo := time.Now().AddDate(-year, -month, -day) // minus 1 месяц
				fileMonthAgoDate := oneMonthAgo.Format("2006.01.02")

				// пробегаем директорию и читаем все файлы, если имя меньше текущее время - месяц = удаляем
				directory, _ := os.Open(l.Dir)
				objects, err := directory.Readdir(-1)
				if err != nil {
					l.Error(err, "Error read directory: ", directory)
					return
				}

				for _, obj := range objects {
					filename := obj.Name()
					filenameMonthAgoDate := fileMonthAgoDate + "_" + l.Service

					if filenameMonthAgoDate > filename {
						pathFile := l.Dir + sep + filename
						err = os.Remove(pathFile)
						if err != nil {
							l.Error(err, "Error deleted file: ", pathFile)
							return
						}
					}
				}
				ticker = time.NewTicker(l.IntervalClearFiles)
			}
		}
	}()
}

// NewFileLogger инициируем логер, которых хранит логи в файлах по указанному пути
func NewFileLogger(ctx context.Context, cfg ConfigLogger) (Log, error) {
	var output io.Writer
	var file *os.File
	var err error
	var mode os.FileMode
	m := sync.Mutex{}

	l := &log{
		Output:             output,
		Levels:             cfg.Level,
		UID:                cfg.Uid,
		Name:               cfg.Name,
		Service:            cfg.Srv,
		Config:             cfg.Config,
		Dir:                cfg.File.Dir,
		IntervalReload:     cfg.File.IntervalReload,
		IntervalClearFiles: cfg.File.IntervalClearFiles,
		PeriodSaveFiles:    cfg.File.PeriodSaveFiles,
		mux:                &m,
		File:               file,
	}

	datefile := time.Now().Format("2006.01.02")
	logName := datefile + "_" + cfg.Srv + "_" + cfg.Uid + ".log"

	fmt.Println(logName)

	// создаем/открываем файл логирования и назначаем его логеру
	mode = 0711
	err = CreateDir(cfg.File.Dir, mode)
	if err != nil {
		logrus.Error(err, "Error creating directory")
		return nil, err
	}

	pathFile := cfg.File.Dir + "/" + logName
	if !IsExist(pathFile) {
		err = CreateFile(pathFile)
		if err != nil {
			logrus.Error(err, "Error creating file")
			return nil, err
		}
	}

	file, err = os.OpenFile(pathFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	l.File = file
	l.Output = file
	if err != nil {
		logrus.Panic(err, "error opening file")
		return nil, err
	}

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			fmt.Printf("panic in loggier (RotateInit). stack: %+v", b)
			//cancel()
			//os.Exit(1)
		}
	}()

	// попытка обновить файл (раз в 10 минут)
	go func() {
		ticker := time.NewTicker(l.IntervalReload)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				l.File.Close() // закрыл старый файл
				datefile = time.Now().Format("2006.01.02")
				logName = datefile + "_" + cfg.Srv + "_" + cfg.Uid + ".log"
				pathFile = cfg.File.Dir + "/" + logName
				if !IsExist(pathFile) {
					err := CreateFile(pathFile)
					if err != nil {
						logrus.Error(err, "Error creating file")
						return
					}
				}

				file, err = os.OpenFile(pathFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					logrus.Panic(err, "error opening file")
					return
				}

				output = file
				l.Output = output
				l.File = file
				ticker = time.NewTicker(l.IntervalReload)
			}
		}
	}()
	l.fileLoggerClearing(ctx)

	return l, err
}
