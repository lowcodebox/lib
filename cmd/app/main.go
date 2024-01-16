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
	applib "git.lowcodeplatform.net/fabric/app/lib"
	iam "git.lowcodeplatform.net/fabric/iam-client"
	"git.lowcodeplatform.net/packages/cache"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"

	implCache "git.lowcodeplatform.net/fabric/app/pkg/cache"
	"git.lowcodeplatform.net/fabric/app/pkg/function"
	"git.lowcodeplatform.net/fabric/app/pkg/i18n"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/servers"
	"git.lowcodeplatform.net/fabric/app/pkg/servers/httpserver"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	"git.lowcodeplatform.net/fabric/app/pkg/session"

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

	err = lib.RunServiceFuncCLI(context.Background(), Start)
	if err != nil {
		fmt.Printf("%s (os.exit 1)", err)
		os.ErrExist = err
		os.Exit(1)
	}

	return
}

// Start стартуем сервис приложения
func Start(ctxm context.Context, configfile, dir, port, mode, proxy, loader, registry, fabric, sourcedb, action, version string) error {
	var cfg model.Config
	var initType string
	var err error

	done := color.Green("[OK]")
	fail := color.Red("[Fail]")

	ctx, cancel := context.WithCancel(ctxm)
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
	cfg.Name, cfg.Version = lib.ValidateNameVersion("", cfg.Type, cfg.Domain)
	cfg.Namespace = cfg.Name + "_" + cfg.Type
	cfg.ClientPath = "/" + cfg.Name + "/" + cfg.Version

	// задаем значение бакера для текущего проекта
	if cfg.VfsBucket == "" {
		cfg.VfsBucket = cfg.Name
	}

	err = logger.SetupDefaultLogboxLogger(cfg.Name+"/"+cfg.Type, logger.LogboxConfig{
		Endpoint:       cfg.LogboxEndpoint,
		AccessKeyID:    cfg.LogboxAccessKeyId,
		SecretKey:      cfg.LogboxSecretKey,
		RequestTimeout: cfg.LogboxRequestTimeout.Value,
		CbMaxRequests:  cfg.CbMaxRequestsLogbox,
		CbTimeout:      cfg.CbTimeoutLogbox.Value,
		CbInterval:     cfg.CbIntervalLogbox.Value,
	}, map[string]string{
		logger.ServiceIDKey:   cfg.HashRun,
		logger.ConfigIDKey:    cfg.UidService,
		logger.ServiceTypeKey: cfg.Type,
	})

	// логируем в консоль, если ошибка подлючения к сервису хранения логов
	if err != nil {
		fmt.Errorf("%s Error init Logbox logger. Was init default logger. err: %s\n", fail, err)
		logger.SetupDefaultLogger(cfg.Name+"/"+cfg.Type,
			logger.WithCustomField(logger.ServiceIDKey, cfg.HashRun),
			logger.WithCustomField(logger.ConfigIDKey, cfg.UidService),
			logger.WithCustomField(logger.ServiceTypeKey, cfg.Type),
		)
	}

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()
	//////////////////////////////////////////////////

	fmt.Printf("%s Enabled logs (type: %s). LogboxEndpoint:%s, Dir:%s\n", done, initType, cfg.LogboxEndpoint, cfg.LogsDir)
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
		ctx,
		cfg.UrlApi,
		cfg.EnableObserverLogApi,
		cfg.CacheRefreshInterval.Value,
		cfg.CbMaxRequests,
		cfg.CbTimeout.Value,
		cfg.CbInterval.Value,
	)

	// инициализация FuncMap
	applib.NewFuncMap(vfs, api)

	// инициализировали переменную кеша
	cache.Init(ctx, 10*time.Hour, 10*time.Minute)

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
		ctx,
		ClearSlash(cfg.UrlIam),
		cfg.ProjectKey,
		cfg.EnableObserverLogIam,
		cfg.CbMaxRequests,
		cfg.CbTimeout.Value,
		cfg.CbInterval.Value,
	)

	ses := session.New(
		ctx,
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

	cache := implCache.New(
		cfg,
		fnc,
	)

	// собираем сервис
	src := service.New(
		ctx,
		cfg,
		cache,
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
		vfs,
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
