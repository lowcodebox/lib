package lib

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Metrics struct {
	StateHost     StateHost
	Connections   int     // количество соединений за весь период учета
	Queue_AVG     float32 // среднее количество запросов в очереди
	Queue_QTL_80  float32 // квантиль 80% - какое среднее кол-во запросов до границы 80% в отсорованном ряду
	Queue_QTL_90  float32 // квантиль 90%
	Queue_QTL_99  float32 // квантиль 99%
	TPR_AVG_MS    float32 // (ms) Time per request - среднее время обработки запроса
	TPR_QTL_MS_80 float32 // (ms) квантиль 80% - какое среднее время обработки запросов до границы 80% в отсорованном ряду
	TPR_QTL_MS_90 float32 // (ms) квантиль 90%
	TPR_QTL_MS_99 float32 // (ms) квантиль 99%

	RPS int // Request per second - количество запросов в секунду
}

type serviceMetric struct {
	Metrics
	Stash          Metrics         // карман для сохранения предыдущего значения
	connectionOpen int             // текущее кол-во открытых соединений (+ при запрос - при ответе)
	queue          []int           // массив соединений в очереди (не закрытых) см.выше
	tpr            []time.Duration // массив времен обработки запросов
	mux            *sync.Mutex
	ctx            context.Context
}

type ServiceMetric interface {
	SetState()
	SetConnectionIncrement()
	SetConnectionDecrement()
	SetTimeRequest(timeRequest time.Duration)
	Generate()
	Get() (result Metrics)
	Clear()
	SaveToStash()
	Middleware(next http.Handler) http.Handler
}

func (s *serviceMetric) SetState() {
	//s.mux.Lock()
	//defer s.mux.Unlock()

	s.StateHost.Tick()

	return
}

// записываем время обработки запроса в массив
func (s *serviceMetric) SetTimeRequest(timeRequest time.Duration) {
	go func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		s.tpr = append(s.tpr, timeRequest)
	}()

	return
}

// SetConnectionIncrement увеличиваем счетчик и добавляем в массив метрик
// формируем временной ряд количества соединений
// при начале запроса увеличиваем, при завершении уменьшаем
// запускаем в отдельной рутине, потому что ф-ция вызывается из сервиса и не должна быть блокирующей
func (s *serviceMetric) SetConnectionIncrement() {
	go func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		s.Connections = s.Connections + 1
		s.connectionOpen = s.connectionOpen + 1
		s.queue = append(s.queue, s.connectionOpen)
	}()

	return
}

// SetConnectionDecrement уменьшаем счетчик и добавляем в массив метрик
// запускаем в отдельной рутине, потому что ф-ция вызывается из сервиса и не должна быть блокирующей
func (s *serviceMetric) SetConnectionDecrement() {
	go func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		if s.connectionOpen != 0 {
			s.connectionOpen = s.connectionOpen - 1
		}
		s.queue = append(s.queue, s.connectionOpen)
	}()

	return
}

func (s *serviceMetric) SetP(value time.Duration) {
	go func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		s.tpr = append(s.tpr, value)
	}()

	return
}

// SaveToStash сохраняем текущее значение расчитанных метрик в кармане
func (s *serviceMetric) SaveToStash() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Stash.StateHost = s.StateHost
	s.Stash.Connections = s.Connections
	s.Stash.RPS = s.RPS

	s.Stash.Queue_AVG = s.Queue_AVG
	s.Stash.Queue_QTL_99 = s.Queue_QTL_99
	s.Stash.Queue_QTL_90 = s.Queue_QTL_90
	s.Stash.Queue_QTL_80 = s.Queue_QTL_80

	s.Stash.TPR_AVG_MS = s.TPR_AVG_MS
	s.Stash.TPR_QTL_MS_80 = s.TPR_QTL_MS_80
	s.Stash.TPR_QTL_MS_90 = s.TPR_QTL_MS_90
	s.Stash.TPR_QTL_MS_99 = s.TPR_QTL_MS_99

	return
}

func (s *serviceMetric) Clear() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Connections = 0
	s.connectionOpen = 0
	s.queue = []int{}
	s.tpr = []time.Duration{}

	s.RPS = 0
	s.Queue_AVG = 0.0
	s.Queue_QTL_80 = 0.0
	s.Queue_QTL_90 = 0.0
	s.Queue_QTL_99 = 0.0

	s.TPR_AVG_MS = 0.0
	s.TPR_QTL_MS_80 = 0.0
	s.TPR_QTL_MS_90 = 0.0
	s.TPR_QTL_MS_99 = 0.0

	return
}

func (s *serviceMetric) Get() (result Metrics) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.Stash
}

func (s *serviceMetric) Generate() {
	var val_Queue_QTL_80, val_Queue_QTL_90, val_Queue_QTL_99, val_Queue float32
	var Queue_AVG, Queue_QTL_80, Queue_QTL_90, Queue_QTL_99 float32
	var val_TPR_80, val_TPR_90, val_TPR_99, val_TPR float32
	var AVG_TPR, QTL_TPR_80, QTL_TPR_90, QTL_TPR_99 float32

	s.mux.Lock()
	defer s.mux.Unlock()

	s.SetState() // БЕЗ БЛОКИРОВКИ получаю текущие метрики загрузки хоста

	//////////////////////////////////////////////////////////
	// расчитываем среднее кол-во запросо и квартили (средние значения после 80-90-99 процентов всех запросов)
	//////////////////////////////////////////////////////////

	// сортируем список
	sort.Ints(s.queue)

	lenQueue := len(s.queue)

	if lenQueue != 0 {
		len_Queue_QTL_80 := lenQueue * 8 / 10
		len_Queue_QTL_90 := lenQueue * 9 / 10
		len_Queue_QTL_99 := lenQueue * 99 / 100

		for i, v := range s.queue {
			vall := float32(v)
			// суммируем значения которые после 80% других
			if i > len_Queue_QTL_80 {
				val_Queue_QTL_80 = val_Queue_QTL_80 + vall
			}
			// суммируем значения которые после 90% других
			if i > len_Queue_QTL_90 {
				val_Queue_QTL_90 = val_Queue_QTL_90 + vall
			}
			// суммируем значения которые после 99% других
			if i > len_Queue_QTL_99 {
				val_Queue_QTL_99 = val_Queue_QTL_99 + vall
			}

			val_Queue = val_Queue + vall
		}

		lQ := float32(lenQueue) - 1 // проверка на 0
		if lQ == 0 {
			lQ = 1
		}
		Queue_AVG = val_Queue / lQ
		Queue_QTL_80 = val_Queue_QTL_80 / float32(lenQueue-len_Queue_QTL_80)
		Queue_QTL_90 = val_Queue_QTL_90 / float32(lenQueue-len_Queue_QTL_90)
		Queue_QTL_99 = val_Queue_QTL_99 / float32(lenQueue-len_Queue_QTL_99)
	}

	//////////////////////////////////////////////////////////
	// расчитываем среднее время запросо и квартили (средние значения после 80-90-99 процентов всех запросов)
	//////////////////////////////////////////////////////////

	// сортируем список
	lenTPR := len(s.tpr)
	if lenTPR != 0 {

		timeInt := []float64{}
		for _, v := range s.tpr {
			timeInt = append(timeInt, float64(v.Microseconds()))
		}
		sort.Float64s(timeInt)

		len_TPR_80 := lenTPR * 8 / 10
		len_TPR_90 := lenTPR * 9 / 10
		len_TPR_99 := lenTPR * 99 / 100

		for i, v := range timeInt {
			vall := float32(v)
			// суммируем значения которые после 80% других
			if i > len_TPR_80 {
				val_TPR_80 = val_TPR_80 + vall
			}
			// суммируем значения которые после 90% других
			if i > len_TPR_90 {
				val_TPR_90 = val_TPR_90 + vall
			}
			// суммируем значения которые после 99% других
			if i > len_TPR_99 {
				val_TPR_99 = val_TPR_99 + vall
			}

			val_TPR = val_TPR + vall
		}

		lQ := float32(lenQueue) - 1
		if lQ == 0 {
			lQ = 1
		}
		AVG_TPR = val_TPR / lQ
		QTL_TPR_80 = val_TPR_80 / float32(lenTPR-len_TPR_80)
		QTL_TPR_90 = val_TPR_90 / float32(lenTPR-len_TPR_90)
		QTL_TPR_99 = val_TPR_99 / float32(lenTPR-len_TPR_99)
	}

	//////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////

	s.RPS = lenQueue / 10

	s.Queue_AVG = Queue_AVG
	s.Queue_QTL_80 = Queue_QTL_80
	s.Queue_QTL_90 = Queue_QTL_90
	s.Queue_QTL_99 = Queue_QTL_99

	s.TPR_AVG_MS = AVG_TPR
	s.TPR_QTL_MS_80 = QTL_TPR_80
	s.TPR_QTL_MS_90 = QTL_TPR_90
	s.TPR_QTL_MS_99 = QTL_TPR_99

	return
}

func (s *serviceMetric) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// увеличиваем счетчик активных сессий
		s.SetConnectionIncrement()
		next.ServeHTTP(w, r)

		// уменьшаем счетчик активных сессий
		s.SetConnectionDecrement()
	})
}

// interval - интервалы времени, через которые статистика будет сбрасыватсья в лог
func NewMetric(ctx context.Context, interval time.Duration) (metrics ServiceMetric) {
	m := sync.Mutex{}
	t := StateHost{}
	s := Metrics{
		StateHost:     t,
		Queue_AVG:     0,
		Queue_QTL_99:  0,
		Queue_QTL_90:  0,
		Queue_QTL_80:  0,
		TPR_AVG_MS:    0,
		TPR_QTL_MS_80: 0,
		TPR_QTL_MS_90: 0,
		TPR_QTL_MS_99: 0,
		RPS:           0,
	}
	metrics = &serviceMetric{
		Metrics:        s,
		Stash:          s,
		connectionOpen: 0,
		queue:          []int{},
		mux:            &m,
		ctx:            ctx,
	}

	go RunMetricLogger(ctx, metrics, interval)

	return metrics
}

func RunMetricLogger(ctx context.Context, m ServiceMetric, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// сохраняем значение метрик в лог
			m.Generate()    // сгенерировали метрики
			m.SaveToStash() // сохранили в карман
			m.Clear()       // очистили объект метрик для приема новых данных
			//mes, _ := json.Marshal(m.Get())
			//logger.Trace(string(mes)) // записали в лог из кармана

			ticker = time.NewTicker(interval)
		}
	}
}
