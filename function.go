package lib

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/araddon/dateparse"
	"github.com/segmentio/ksuid"
	duration "github.com/xhit/go-str2duration"
)

var (
	reDate     = regexp.MustCompile(`^(\d{2})[./](\d{2})[./](\d{4})\b`)
	reInterval = regexp.MustCompile(`(.+) ([+-]) (\d[\d.wdhms]*[wdhms])$`)
	reUTC      = regexp.MustCompile(`(?i)\b(?:UTC|GMT)([+-])(\d+)$`)

	LocationMSK = time.FixedZone("Europe/Moscow", 3*3600)
)

// ResponseJSON если status не из списка, то вставляем статус - 501 и Descraption из статуса
func ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) (err error) {

	if w == nil {
		return
	}

	errMessage := models.RestStatus{}
	st, found := models.StatusCode[status]
	if found {
		errMessage = st
	} else {
		errMessage = models.StatusCode["NotStatus"]
	}

	objResp := &models.Response{}
	if error != nil {
		errMessage.Error = error
	}

	// Metrics
	b1, _ := json.Marshal(metrics)
	var metricsR models.Metrics
	json.Unmarshal(b1, &metricsR)
	if metrics != nil {
		objResp.Metrics = metricsR
	}

	objResp.Status = errMessage
	objResp.Data = objResponse

	// формируем ответ
	out, err := json.Marshal(objResp)
	if err != nil {
		out = []byte(fmt.Sprintf("%s", err))
	}

	//WriteFile("./dump.json", out)
	w.WriteHeader(errMessage.Status)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(out)

	return
}

// RunProcess стартуем сервис из конфига
func RunProcess(path, config, command, mode, dc string) (pid int, err error) {
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if config == "" {
		return 0, errors.New("config file not specified")
	}
	if command == "" {
		command = "start"
	}

	path = strings.Replace(path, "//", "/", -1)

	cmd = exec.Command(path, command, "--config", config, "--mode", mode, "--dc", dc)
	if mode == "debug" {
		t := time.Now().Format("2006.01.02-15-04-05")
		s := strings.Split(path, sep)
		srv := s[len(s)-1]

		dirPath := "debug" + sep + srv
		err = CreateDir(dirPath, 0777)
		if err != nil {
			return 0, fmt.Errorf("unable create directory for debug file, path: %s, err: %w", dirPath, err)
		}

		filePath := "debug" + sep + srv + sep + fmt.Sprint(t) + "_" + UUID()[:6] + ".log"
		f, err := os.Create(filePath)
		if err != nil {
			return 0, fmt.Errorf("unable create debug file, path: %s, err: %w", filePath, err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("unable start process, status: %d, config: %s, path: %s, command: %s, mode: %s, dc: %s, err: %w",
			cmd.ProcessState.ExitCode(), config, path, command, mode, dc, err)
	}

	go cmd.Process.Wait()

	pid = cmd.Process.Pid

	// в течение заданного интервала ожидаем завершающий статус запуска
	// или выходим если -1 (в процессе или прибит сигналом)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		exitCode := cmd.ProcessState.ExitCode()

		// завершился
		if exitCode >= 0 {
			return
		}

		select {
		case <-ctx.Done():
			return

		default:
			// -1 — работает или прибит сигналом
		}
	}

	return
}

// RootDir получаем корневую директорию от места где запускаем файл
func RootDir() (rootDir string, err error) {
	file, err := filepath.Abs(os.Args[0])
	if err != nil {
		return
	}
	rootDir = path.Dir(file)
	if err != nil {
		fmt.Println("Error calculation RootDir. File: ", file, "; Error: ", err)
	}

	return
}

func Hash(str string) (result string) {
	h := sha1.New()
	h.Write([]byte(str))
	result = hex.EncodeToString(h.Sum(nil))

	return
}

func PanicOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}
}

func UUID() (result string) {
	return ksuid.New().String()
}

// RemoveElementFromData удаляем элемент из слайса
func RemoveElementFromData(p *models.ResponseData, i int) bool {

	if i < len(p.Data) {
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}

// JsonEscape экранируем "
// fmt.Println(jsonEscape(`dog "fish" cat`))
// output: dog \"fish\" cat
func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// SearchConfigDir получаем путь до искомой конфигурации от переданной директории
func SearchConfig(projectDir, configuration string) (configPath string, err error) {
	var nextPath string

	directory, err := os.Open(projectDir)
	if err != nil {
		return "", err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return "", err
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range objects {

		nextPath = projectDir + sep + obj.Name()
		if obj.IsDir() {
			dirName := obj.Name()

			// не входим в скрытые папки
			if dirName[:1] != "." {
				configPath, err = SearchConfig(nextPath, configuration)
				if configPath != "" {
					return configPath, err // поднимает результат наверх
				}
			}
		} else {
			if !strings.Contains(nextPath, "/.") {
				// проверяем только файлы конфигурации (игнорируем .json)
				if strings.Contains(obj.Name(), configuration+".cfg") {
					return nextPath, err
				}
			}
		}
	}

	return configPath, err
}

// TimeParse парсит дату-время из любого формата.
// Если в строке не передана временна́я зона, то парсится как UTC.
//
// Если вторым параметром передать true, то полученное время скастится в UTC.
//
// Можно задать интервалы, которые надо добавить/вычесть, знак операции при этом отбивается пробелами.
func TimeParse(str string, toUTC bool) (res time.Time, err error) {
	var (
		signs, intervals []string
		dur              time.Duration
	)

	// извлекаем интервал из строки при наличии
	for {
		if reInterval.MatchString(str) {
			signs = append(signs, reInterval.ReplaceAllString(str, "$2"))
			intervals = append(intervals, reInterval.ReplaceAllString(str, "$3"))
			str = reInterval.ReplaceAllString(str, "$1")
		} else {
			break
		}
	}

	// приводим дату к стандартному формату
	str = reDate.ReplaceAllString(str, "$3-$2-$1")

	// делаем часовой пояс понятнее парсеру
	if reUTC.MatchString(str) {
		if zone := reUTC.FindStringSubmatch(str); len(zone) >= 2 && len(zone[2]) < 4 {
			shift := zone[2]
			if len(shift) == 1 {
				shift = "0" + shift
			}
			if len(shift) == 2 {
				shift += "00"
			}
			str = reUTC.ReplaceAllString(str, zone[1]+shift)
		}
	}

	res, err = dateparse.ParseAny(str)
	if err != nil {
		return
	}

	// смещаем на заданный интервал
	for i, interval := range intervals {
		dur, err = duration.Str2Duration(interval)
		if err != nil {
			return
		}

		if signs[i] == "-" {
			dur = -dur
		}
		res = res.Add(dur)
	}

	if toUTC {
		return res.UTC(), nil
	}

	return res, nil
}

// Принимает на вход объект, текущую роль пользака и режим ("read", "write", "delete", "admin")
func CheckRoles(obj models.Data, role string, mode string) (bool, error) {
	access_read, foundRead := obj.Attr("access_read", "src")
	access_write, foundWrite := obj.Attr("access_write", "src")
	access_delete, foundDelete := obj.Attr("access_delete", "src")
	access_admin, foundAdmin := obj.Attr("access_admin", "src")

	if access_read == "" && access_write == "" && access_delete == "" && access_admin == "" {
		return true, nil
	} else {
		switch mode {
		case "read":
			if !foundRead {
				return false, errors.New("error not found access_read attr")
			} else if strings.Contains(access_read, role) {
				return true, nil
			} else {
				return false, nil
			}
		case "write":
			if !foundWrite {
				return false, errors.New("error not found access_write attr")
			} else if strings.Contains(access_write, role) {
				return true, nil
			} else {
				return false, nil
			}
		case "delete":
			if !foundDelete {
				return false, errors.New("error not found access_delete attr")
			} else if strings.Contains(access_delete, role) {
				return true, nil
			} else {
				return false, nil
			}
		case "admin":
			if !foundAdmin {
				return false, errors.New("error not found access_admin attr")
			} else if strings.Contains(access_admin, role) {
				return true, nil
			} else {
				return false, nil
			}
		default:
			return false, errors.New("error unknown mode")
		}
	}
}
