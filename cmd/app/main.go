package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/api-client"
	iam "git.edtech.vm.prod-6.cloud.el/fabric/iam-client"
	"git.edtech.vm.prod-6.cloud.el/packages/cache"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	applib "git.edtech.vm.prod-6.cloud.el/fabric/app/lib"

	implCache "git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/cache"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/function"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/i18n"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/servers"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/servers/httpserver"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/service"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/session"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/lovetsky/secrets"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	analytics "git.edtech.vm.prod-6.cloud.el/wb/analyticscollector-client"
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

	ctx, cancel := signal.NotifyContext(ctxm, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// инициируем пакеты
	cfgString, err := lib.ConfigLoad(configfile, &cfg)
	if err != nil {
		return fmt.Errorf("%s (%s)", "Error. Load config is failed.", err)
	}
	err = secrets.ParseSecrets(context.Background(), cfgString, cfg.ProxyPointsrc, cfg.ProjectKey, &cfg)
	if err != nil {
		return fmt.Errorf("error. Parse secrets is failed. (%w)", err)
	}

	cfg.ServiceVersion = serviceVersion
	cfg.HashCommit = hashCommit
	cfg.UidService = cfg.DataUid
	cfg.ConfigName = cfg.DataUid
	cfg.HashRun = lib.UUID()
	cfg.Name, cfg.Version = lib.ValidateNameVersion("", cfg.Type, cfg.Domain)
	cfg.Namespace = cfg.Name + "_" + cfg.Type
	cfg.ClientPath = "/" + cfg.Name + "/" + cfg.Version
	cfg.HashRun = lib.UUID()
	cfg.RunTime = time.Now()
	cfg.Pid = strconv.Itoa(os.Getpid())

	if cfg.Environment == "" {
		cfg.Environment = cfg.EnvironmentPointsrc
	}
	if cfg.Cluster == "" {
		cfg.Cluster = cfg.ClusterPointsrc
	}

	go func() {
		for {
			cfg.UpTime = time.Now().Sub(cfg.RunTime).String()
			time.Sleep(5 * time.Second)
		}
	}()

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

	time.Sleep(5 * time.Second)

	fmt.Printf("%s Enabled logs (type: %s). LogboxEndpoint:%s, Dir:%s\n", done, initType, cfg.LogboxEndpoint, cfg.LogsDir)
	logger.Info(ctx, "Запускаем app-сервис", zap.String("domain", cfg.Domain))
	//////////////////////////////////////////////////

	// создаем метрики
	//metrics := lib.NewMetric(
	//	ctx,
	//	logger,
	//	cfg.LogIntervalMetric.Value,
	//)

	//defer func() {
	//	rec := recover()
	//	if rec != nil {
	//		b := string(debug.Stack())
	//		logger.Panic(ctx, "Recover panic from main function.", zap.String("debug stack", b))
	//		cancel()
	//		runtime.Goexit()
	//	}
	//}()

	// клиент аналитики
	analyticsClient, err := analytics.New(ctx, cfg.AnalyticsHost, cfg.LogboxRequestTimeout.Value, "/")
	if err != nil {
		fmt.Printf("%s Can't create analitics client\nHost: %s err: %s", fail, cfg.AnalyticsHost, err)
		logger.Info(ctx, "can't create analitics client", zap.String("host", cfg.AnalyticsHost), zap.Error(err))
		return err
	}

	msg := i18n.New()

	api := api.New(
		ctx,
		cfg.UrlApi,
		cfg.EnableObserverLogApi,
		cfg.CacheRefreshInterval.Value,
		cfg.CbMaxRequests,
		cfg.CbTimeout.Value,
		cfg.CbInterval.Value,
		cfg.ProjectKey,
	)

	//fmt.Println(api.ObjCreate(ctx, map[string]string{"data-uid": "test-2024-02-08T08-27-09Z-2c4piB"}))
	fmt.Printf("%s Enabled API (url: %s)\n", done, cfg.UrlApi)

	// инициализация FuncMap
	applib.NewFuncMap(vfs, api, &cfg, cfg.ProjectKey, analyticsClient)

	// инициализировали переменную кеша
	cache.Init(ctx, 10*time.Hour, 10*time.Minute)

	fnc := function.New(
		cfg,
		api,
	)

	if strings.TrimSuffix(cfg.UrlIam, "/") == "" {
		logger.Error(ctx, "Error: UrlIam is empty", zap.Error(err))
		fmt.Println("Error: UrlIam is empty")
		return err
	}

	iam := iam.New(
		ctx,
		strings.TrimSuffix(cfg.UrlIam, "/"),
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
		port, err = lib.ProxyPort(cfg.ProxyPointsrc, cfg.PortInterval, cfg.ProxyMaxCountRetries.Value, cfg.ProxyTimeRetries.Value)
		if err != nil {
			logger.Error(context.Background(), "port is not resolved", zap.String("proxyPath", cfg.ProxyPointsrc),
				zap.String("portInterval", cfg.PortInterval), zap.String("proxyPath", cfg.ProxyPointsrc))
			err = fmt.Errorf("port is not resolved")
			fmt.Printf("%s Port is not resolved! path: %s, interval: %s\n", fail, cfg.ProxyPointsrc, cfg.PortInterval)
			return err
		}
	}
	cfg.PortApp = port

	cache, err := implCache.New(
		cfg,
		fnc,
	)
	if err == nil {
		fmt.Printf("%s Cache-service is running\n", done)
		logger.Info(ctx, "cache is running")
	} else {
		fmt.Printf("%s Cache-service is not running\n (cfg.CachePointsrc: %s, err: %s)\n", fail, cfg.CachePointsrc, err)
		logger.Info(ctx, "cache running is failed", zap.String("CachePointsrc", cfg.CachePointsrc), zap.Error(err))
	}

	fmt.Printf("%s Analitics client connected", done)
	logger.Info(ctx, "analitics client connected")

	// собираем сервис
	src := service.New(
		ctx,
		&cfg,
		cache,
		msg,
		ses,
		api,
		iam,
		vfs,
	)

	//TODO: сделать норм инициализацию, сделано так чтобы использовать DogParse()
	app_lib := applib.New(
		nil,
		"",
		"",
		"",
		nil,
		"",
		"",
		nil,
		vfs,
		api,
	)

	// httpserver
	httpserver := httpserver.New(
		ctx,
		cfg,
		src,
		iam,
		ses,
		vfs,
		api,
		app_lib,

		serviceVersion,
		hashCommit,
	)

	srv := servers.New(
		"http",
		src,
		httpserver,
		cfg,
	)

	errChannel := make(chan error)
	go func() {
		errChannel <- srv.Run()
	}()

	select {
	case <-ctx.Done():
		close(errChannel)
		errs := make([]error, len(errChannel))
		for err := range errChannel {
			errs = append(errs, err)
		}
		return multierr.Combine(errs...)
	case err := <-errChannel:
		return err
	}
}
