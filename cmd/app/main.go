package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
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
	"go.uber.org/zap"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/packages/logger"
)

const sep = string(os.PathSeparator)
const prefixUploadURL = "upload" // адрес/_prefixUploadURL_/... - путь, относительно bucket-а проекта
var (
	serviceVersion string
	hashCommit     string
)

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

	var err error

	err = lib.RunServiceFuncCLI(Start)
	if err != nil {
		fmt.Printf("%s (os.exit 1)", err)
		os.ErrExist = err
		os.Exit(1)
	}

	return
}

// Start стартуем сервис приложения
func Start(configfile, dir, port, mode, proxy, loader, registry, fabric, sourcedb, action, version string) error {
	var cfg model.Config
	var initType string
	var err error

	done := color.Green("[OK]")
	fail := color.Red("[Fail]")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// инициируем пакеты
	err = lib.ConfigLoad(configfile, &cfg)
	if err != nil {
		return fmt.Errorf("%s (%s)", "Error. Load config is failed.", err)
	}

	cfg.ServiceVersion = serviceVersion
	cfg.HashCommit = hashCommit
	cfg.UidService = cfg.DataUid
	cfg.ConfigName = cfg.DataUid
	cfg.HashRun = lib.UUID()
	cfg.Name, cfg.ServiceType = lib.ValidateNameVersion("", cfg.ServiceType, cfg.Domain)

	// задаем значение бакера для текущего проекта
	if cfg.VfsBucket == "" {
		cfg.VfsBucket = cfg.Name
	}

	err = logger.SetupDefaultLogboxLogger(cfg.Name+"/"+cfg.ServiceType, logger.LogboxConfig{
		Endpoint:       cfg.LogboxEndpoint,
		AccessKeyID:    cfg.LogboxAccessKeyId,
		SecretKey:      cfg.LogboxSecretKey,
		RequestTimeout: cfg.LogboxRequestTimeout.Value,
	}, map[string]string{
		logger.ServiceIDKey:   cfg.HashRun,
		logger.ConfigIDKey:    cfg.UidService,
		logger.ServiceTypeKey: cfg.ServiceType,
	})

	// логируем в консоль, если ошибка подлючения к сервису хранения логов
	if err != nil {
		fmt.Errorf("%s Error init Logbox logger. Was init default logger. err: %s\n", fail, err)
		logger.SetupDefaultLogger(cfg.Name+"/"+cfg.ServiceType,
			logger.WithCustomField(logger.ServiceIDKey, cfg.HashRun),
			logger.WithCustomField(logger.ConfigIDKey, cfg.UidService),
			logger.WithCustomField(logger.ServiceTypeKey, cfg.ServiceType),
		)
	}

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)
	err = vfs.Connect()
	if err != nil {
		logger.Error(ctx, "Error connect to filestorage", zap.String("configfile", configfile), zap.Error(err))
		return fmt.Errorf("%s Error connect to filestorage. err: %s\n cfg: VfsKind: %s, VfsEndpoint: %s, VfsAccessKeyId: %s, VfsSecretKey: %s, VfsRegion: %s, VfsBucket: %s, VfsComma: %s", fail, err, cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma)
	}

	fmt.Printf("%s Enabled logs (type: %s). Level:%s, Dir:%s\n", done, initType, cfg.LogsLevelPointsrc, cfg.LogsDir)
	logger.Info(ctx, "Запускаем app-сервис: ", zap.String("domain", cfg.Domain))

	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()
	//////////////////////////////////////////////////

	fmt.Printf("%s Enabled logs (type: %s). Level:%s, Dir:%s\n", done, initType, cfg.LogsLevelPointsrc, cfg.LogsDir)
	logger.Info(ctx, "Запускаем app-сервис: ", zap.String("domain", cfg.Domain))
	//////////////////////////////////////////////////

	// создаем метрики
	//metrics := lib.NewMetric(
	//	ctx,
	//	logger,
	//	cfg.LogIntervalMetric.Value,
	//)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(ctx, "Recover panic from main function.", zap.String("debug stack", b))
			cancel()
			runtime.Goexit()
		}
	}()

	msg := i18n.New()

	api := api.New(
		cfg.UrlApi,
	)

	fnc := function.New(
		cfg,
		api,
	)

	if ClearSlash(cfg.UrlIam) == "" {
		logger.Error(ctx, "Error: UrlIam is empty", zap.Error(err))
		fmt.Println("Error: UrlIam is empty")
		return err
	}

	iam := iam.New(
		ClearSlash(cfg.UrlIam),
		cfg.ProjectKey,
		nil,
		nil,
	)

	ses := session.New(
		cfg,
		api,
		iam,
	)

	// запускаем очиститель сессий
	go ses.Cleaner(ctx)

	if port == "" {
		port, err = lib.AddressProxy(cfg.ProxyPointsrc, cfg.PortInterval)
		if err != nil {
			logger.Error(ctx, "Error: AddressProxy", zap.String("ProxyPointsrc", cfg.ProxyPointsrc), zap.String("ProxyPointsrc", cfg.ProxyPointsrc), zap.Error(err))

			fmt.Println(err)
			return err
		}
	}
	cfg.PortApp = port

	cach := cache.New(
		cfg,
		fnc,
	)

	// собираем сервис
	src := service.New(
		ctx,
		cfg,
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
		iam,
		ses,
		serviceVersion,
		hashCommit,
	)

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go ListenForShutdown(ch, cancel)

	srv := servers.New(
		"http",
		src,
		httpserver,
		cfg,
	)
	srv.Run()

	return err
}

func ListenForShutdown(ch <-chan os.Signal, cancelFunc context.CancelFunc) {
	var done = color.Grey("[OK]")
	ctx := context.Background()

	<-ch
	cancelFunc()
	logger.Info(ctx, "Service is stopped. Logfile is closed.")

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
