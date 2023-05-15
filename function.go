package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"git.lowcodeplatform.net/fabric/models"
	uuid "github.com/satori/go.uuid"
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
func RunProcess(path, config, command, mode string) (pid int, err error) {
	var cmd *exec.Cmd

	if config == "" {
		return 0, fmt.Errorf("%s", "Configuration file is not found")
	}
	if command == "" {
		command = "start"
	}

	path = strings.Replace(path, "//", "/", -1)

	cmd = exec.Command(path, command, "--config", config, "--mode", mode)
	if mode == "debug" {
		t := time.Now().Format("2006.01.02-15-04-05")
		s := strings.Split(path, sep)
		srv := s[len(s)-1]

		dirPath := "debug" + sep + srv
		err = CreateDir(dirPath, 0777)
		if err != nil {
			return 0, fmt.Errorf("error create directory for debug-file. path: %s, err: %s", dirPath, err)
		}

		filePath := "debug" + sep + srv + sep + fmt.Sprint(t) + "_" + UUID()[:6] + ".log"
		f, err := os.Create(filePath)
		if err != nil {
			return 0, fmt.Errorf("error create debug-file. path: %s, err: %s", filePath, err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("status: %d, config: %s", cmd.ProcessState.ExitCode(), config)

		return 0, err
	}

	pid = cmd.Process.Pid

	time.Sleep(10 * time.Second)
	if cmd.ProcessState.ExitCode() != 0 {
		err = fmt.Errorf("status: %d, config: %s, err: %s", cmd.ProcessState.ExitCode(), config, cmd.Stderr)
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
	stUUID := uuid.NewV4()
	return stUUID.String()
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
