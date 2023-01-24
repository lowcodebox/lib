package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"git.lowcodeplatform.net/fabric/api-client"
	"git.lowcodeplatform.net/fabric/app/pkg/cache"
	"git.lowcodeplatform.net/fabric/app/pkg/function"
	"git.lowcodeplatform.net/fabric/app/pkg/i18n"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/servers"
	"git.lowcodeplatform.net/fabric/app/pkg/servers/httpserver"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	"git.lowcodeplatform.net/fabric/app/pkg/session"
	iam "git.lowcodeplatform.net/fabric/iam-client"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"

	"git.lowcodeplatform.net/fabric/lib"
)

const sep = string(os.PathSeparator)
const prefixUploadURL = "upload" // адрес/_prefixUploadURL_/... - путь, относительно bucket-а проекта

func main() {
	//limit := 1
	//burst := 1
	//limiter := rate.NewLimiter(rate.Limit(limit), burst)
	//ctx := context.Background()
	//i := 0
	//
	//for {
	//
	//	fmt.Println("request", time.Now())
	//
	//	go func(lim *rate.Limiter, i int) {
	//		fmt.Println(i, "------- - ", time.Now())
	//		lim.Wait(ctx)
	//		fmt.Println(i, " - ", time.Now())
	//	}(limiter, i)
	//
	//	i++
	//
	//	time.Sleep(100 * time.Millisecond)
	//	if i > 20 {
	//		break
	//	}
	//}

	//time.Sleep(100 * time.Second)

	lib.RunServiceFuncCLI(Start)
}

// Start стартуем сервис приложения
func Start(configfile, dir, port, mode, proxy, loader, registry, fabric, sourcedb, action, version string) {
	var cfg model.Config

	done := color.Green("[OK]")
	fail := color.Red("[Fail]")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// инициируем пакеты
	err := lib.ConfigLoad(configfile, &cfg)
	if err != nil {
		fmt.Printf("%s (%s)", "Error. Load config is failed.", err)
		return
	}

	cfg.UidService = cfg.DataUid

	// подключаемся к файловому хранилищу
	if cfg.VfsBucket == "" {
		cfg.VfsBucket = strings.Replace(cfg.Domain, "/", "_", -1)
	}
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma)
	err = vfs.Connect()
	if err != nil {
		fmt.Printf("%s Error connect to filestorage: %s\n", fail, err)
		log.Info("Error connect to filestorage (", configfile, ")", err)
		return
	}

	// формируем значение переменных по-умолчанию или исходя из требований сервиса
	cfg.SetClientPath()
	cfg.SetRootDir()
	cfg.SetConfigName()

	cfg.Workingdir = cfg.RootDir

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// формирование пути к лог-файлам и метрикам
	if cfg.LogsDir == "" {
		cfg.LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if cfg.LogsDir[:1] != sep {
		rootDir, _ := lib.RootDir()
		cfg.LogsDir = rootDir + sep + "upload" + sep + cfg.Domain + sep + cfg.LogsDir
	}
	// инициализируем кеширование
	cfg.Namespace = strings.ReplaceAll(cfg.Domain, "/", "_")
	cfg.UrlProxy = cfg.ProxyPointsrc

	// инициализировать лог и его ротацию
	var logger = lib.NewLogger(
		cfg.LogsDir,
		cfg.LogsLevel,
		lib.UUID(),
		cfg.Domain,
		"app",
		cfg.UidService,
		cfg.LogIntervalReload.Value,
		cfg.LogIntervalClearFiles.Value,
		cfg.LogPeriodSaveFiles,
	)
	logger.RotateInit(ctx)

	fmt.Printf("%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
	logger.Info("Запускаем app-сервис: ", cfg.Domain)

	// создаем метрики
	metrics := lib.NewMetric(
		ctx,
		logger,
		cfg.LogIntervalMetric.Value,
	)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
			cancel()
			runtime.Goexit()
		}
	}()

	msg := i18n.New()

	api := api.New(
		cfg.UrlApi,
		logger,
		metrics,
	)

	fnc := function.New(
		cfg,
		logger,
		api,
	)

	if ClearSlash(cfg.UrlIam) == "" {
		fmt.Println("Error: UrlIam is empty")
		return
	}

	iam := iam.New(
		ClearSlash(cfg.UrlIam),
		cfg.ProjectKey,
		logger,
		metrics,
	)

	ses := session.New(
		logger,
		cfg,
		api,
		iam,
	)

	// запускаем очиститель сессий
	go ses.Cleaner(ctx)

	if port == "" {
		port, err = lib.AddressProxy(cfg.UrlProxy, cfg.PortInterval)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	cfg.PortApp = port

	cach := cache.New(
		cfg,
		logger,
		fnc,
	)

	// собираем сервис
	src := service.New(
		ctx,
		logger,
		cfg,
		metrics,
		cach,
		msg,
		ses,
		api,
		iam,
		vfs,
	)

	// httpserver
	httpserver := httpserver.New(
		ctx,
		cfg,
		src,
		metrics,
		logger,
		iam,
		ses,
	)

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go ListenForShutdown(ch, logger, cancel)

	srv := servers.New(
		"http",
		src,
		httpserver,
		metrics,
		cfg,
	)
	srv.Run()
}

func ListenForShutdown(ch <-chan os.Signal, logger lib.Log, cancelFunc context.CancelFunc) {
	var done = color.Grey("[OK]")

	<-ch
	cancelFunc()
	logger.Warning("Service is stopped. Logfile is closed.")
	logger.Close()

	fmt.Printf("%s Service is stopped. Logfile is closed.\n", done)

	time.Sleep(2 * time.Second)

	os.Exit(0)
}

func ClearSlash(str string) (result string) {
	if len(str) > 1 {
		if str[len(str)-1:] == "/" {
			result = str[:len(str)-1]
		} else {
			result = str
		}
	}
	return result
}
