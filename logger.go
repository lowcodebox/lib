// обертка для логирования, которая дополняем аттрибутами логируемого процесса logrus
// дополняем значениями, идентифицирующими запущенный сервис UID,Name,Service

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

var logrusB = logrus.New()

// LogLine структура строк лог-файла. нужна для анмаршалинга
type LogLine struct {
	Config string      `json:"config"`
	Level  string      `json:"level"`
	Msg    interface{} `json:"msg"`
	Name   string      `json:"name"`
	Srv    string      `json:"srv"`
	Time   string      `json:"time"`
	Uid    string      `json:"uid"`
}

type log struct {

	// куда логируем? stdout/;*os.File на файл, в который будем писать логи
	Output io.Writer `json:"output"`
	//Debug:
	// сообщения отладки, профилирования.
	// В production системе обычно сообщения этого уровня включаются при первоначальном
	// запуске системы или для поиска узких мест (bottleneck-ов).

	//Info: - логировать процесс выполнения
	// обычные сообщения, информирующие о действиях системы.
	// Реагировать на такие сообщения вообще не надо, но они могут помочь, например,
	// при поиске багов, расследовании интересных ситуаций итд.

	//Warning: - логировать странные операции
	// записывая такое сообщение, система пытается привлечь внимание обслуживающего персонала.
	// Произошло что-то странное. Возможно, это новый тип ситуации, ещё не известный системе.
	// Следует разобраться в том, что произошло, что это означает, и отнести ситуацию либо к
	// инфо-сообщению, либо к ошибке. Соответственно, придётся доработать код обработки таких ситуаций.

	//Error: - логировать ошибки
	// ошибка в работе системы, требующая вмешательства. Что-то не сохранилось, что-то отвалилось.
	// Необходимо принимать меры довольно быстро! Ошибки этого уровня и выше требуют немедленной записи в лог,
	// чтобы ускорить реакцию на них. Нужно понимать, что ошибка пользователя – это не ошибка системы.
	// Если пользователь ввёл в поле -1, где это не предполагалось – не надо писать об этом в лог ошибок.

	//Panic: - логировать критические ошибки
	// это особый класс ошибок. Такие ошибки приводят к неработоспособности системы в целом, или
	// неработоспособности одной из подсистем. Чаще всего случаются фатальные ошибки из-за неверной конфигурации
	// или отказов оборудования. Требуют срочной, немедленной реакции. Возможно, следует предусмотреть уведомление о таких ошибках по SMS.
	// указываем уровни логирования Error/Warning/Debug/Info/Panic

	//Trace: - логировать обработки запросов

	// можно указывать через | разные уровени логирования, например Error|Warning
	// можно указать All - логирование всех уровней
	Levels string `json:"levels"`
	// uid процесса (сервиса), который логируется (случайная величина)
	UID string `json:"uid"`
	// имя процесса (сервиса), который логируется
	Name string `json:"name"`
	// название сервиса (app/gui...)
	Service string `json:"service"`
	// директория сохранения логов
	Dir string `json:"dir"`
	// uid-конфигурации с которой был запущен процесс
	Config string `json:"config"`
	// интервал между проверками актуального файла логирования (для текущего дня)
	IntervalReload time.Duration `json:"delay_reload"`
	// интервал проверками на наличие файлов на удаление
	IntervalClearFiles time.Duration `json:"interval_clear_files"`
	// период хранения файлов лет-месяцев-дней (например: 0-1-0 - хранить 1 месяц)
	PeriodSaveFiles string `json:"period_save_files"`

	File *os.File

	mux *sync.Mutex
}

type Log interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(err error, args ...interface{})
	Panic(err error, args ...interface{})
	Exit(err error, args ...interface{})
	RotateInit(ctx context.Context)
	GetOutput() io.Writer
	GetFile() *os.File
	Close()
}

func (l *log) Trace(args ...interface{}) {
	if strings.Contains(l.Levels, "Trace") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.TraceLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
		}).Trace(args...)
	}
}

func (l *log) Debug(args ...interface{}) {
	if strings.Contains(l.Levels, "Debug") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		// Only log the warning severity or above.
		logrusB.SetLevel(logrus.DebugLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
		}).Debug(args...)
	}
}

func (l *log) Info(args ...interface{}) {
	if strings.Contains(l.Levels, "Info") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.SetLevel(logrus.InfoLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
		}).Info(args...)
	}
}

func (l *log) Warning(args ...interface{}) {
	if strings.Contains(l.Levels, "Warning") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.WarnLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
		}).Warn(args...)
	}
}

func (l *log) Error(err error, args ...interface{}) {
	if strings.Contains(l.Levels, "Error") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.ErrorLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
			"error":  fmt.Sprint(err),
		}).Error(args...)
	}
}

func (l *log) Panic(err error, args ...interface{}) {
	if strings.Contains(l.Levels, "Panic") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.PanicLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
			"error":  fmt.Sprint(err),
		}).Panic(args...)
	}
}

// Exit внутренняя ф-ция логирования и прекращения работы программы
func (l *log) Exit(err error, args ...interface{}) {
	if strings.Contains(l.Levels, "Fatal") {
		logrusB.SetOutput(l.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.FatalLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   l.Name,
			"uid":    l.UID,
			"srv":    l.Service,
			"config": l.Config,
			"error":  fmt.Sprint(err),
		}).Fatal(args...)
	}
}

// RotateInit Переинициализация файла логирования
func (l *log) RotateInit(ctx context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l.IntervalReload = 5 * time.Second

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
				b := NewLogger(l.Dir, l.Levels, l.UID, l.Name, l.Service, l.Config, l.IntervalReload, l.IntervalClearFiles, l.PeriodSaveFiles)

				l.Output = b.GetOutput()
				l.File = b.GetFile() // передал указатель на новый файл в структуру лога
				ticker = time.NewTicker(l.IntervalReload)
			}
		}
	}()

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

func (l *log) GetOutput() io.Writer {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.Output
}

func (l *log) GetFile() *os.File {

	return l.File
}

func (l *log) Close() {
	l.File.Close()
}

func NewLogger(logsDir, level, uid, name, srv, config string, intervalReload, intervalClearFiles time.Duration, periodSaveFiles string) Log {
	var output io.Writer
	var file *os.File
	var err error
	var mode os.FileMode
	m := sync.Mutex{}

	datefile := time.Now().Format("2006.01.02")
	logName := datefile + "_" + srv + "_" + uid + ".log"

	// создаем/открываем файл логирования и назначаем его логеру
	mode = 0711
	CreateDir(logsDir, mode)
	if err != nil {
		logrus.Error(err, "Error creating directory")
		return nil
	}

	pathFile := logsDir + "/" + logName

	if !IsExist(pathFile) {
		err := CreateFile(pathFile)
		if err != nil {
			logrus.Error(err, "Error creating file")
			return nil
		}
	}

	file, err = os.OpenFile(pathFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	output = file
	if err != nil {
		logrus.Panic(err, "error opening file")
		return nil
	}

	return &log{
		Output:             output,
		Levels:             level,
		UID:                uid,
		Name:               name,
		Service:            srv,
		Dir:                logsDir,
		Config:             config,
		IntervalReload:     intervalReload,
		IntervalClearFiles: intervalClearFiles,
		PeriodSaveFiles:    periodSaveFiles,
		mux:                &m,
		File:               file,
	}
}
