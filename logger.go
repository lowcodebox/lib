// обертка для логирования, которая дополняем аттрибутами логируемого процесса logrus
// дополняем значениями, идентифицирующими запущенный сервис UID,Name,Service

package lib

import (
	"fmt"
	"io"
	"os"
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

	// путь к сервису отправки логов в хранилище (Logbox)
	LogboxURL string
	// интервал отправки (в промежутках сохраняем в буфер)
	LogboxSendInterval time.Duration

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

func (l *log) Close() {
	l.File.Close()
}