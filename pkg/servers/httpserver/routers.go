package httpserver

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/servers/httpserver/handlers"
)

type Result struct {
	Status  string `json:"status"`
	Content []interface{}
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (h *httpserver) NewRouter(checkHttpsOnly bool) (*mux.Router, error) {
	router := mux.NewRouter().StrictSlash(true)
	handler := handlers.New(h.src, h.cfg, h.api, h.vfs, h.app_lib)

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	proxy, err := h.vfs.Proxy("/upload", "/"+h.cfg.VfsBucket)
	if err != nil {
		logger.Panic(h.ctx, "unable init s3 proxy", zap.Error(err))
	}

	router.Handle("/upload/{params:.+}", proxy).Methods(http.MethodGet)
	router.Name("Metrics").Path("/metrics").Handler(promhttp.Handler())
	router.Name("Storage").Path("/assets/{params:.+}").Methods(http.MethodGet).HandlerFunc(handler.Storage)
	router.Name("Storage").Path("/templates/{params:.+}").Methods(http.MethodGet).HandlerFunc(handler.Storage)

	//prometheus.MustRegister(version.NewCollector(h.cfg.Name))
	version.Version = h.serviceVersion
	version.Revision = h.hashCommit

	//_, err = cache.Cache().Upsert("prometheus", func() (res interface{}, err error) {
	//	mf, err := prometheus.DefaultGatherer.Gather()
	//	if err != nil {
	//		err = fmt.Errorf("error prometheus Gather. err: %s", err)
	//		return
	//	}
	//	res, err = json.Marshal(mf)
	//	return res, nil
	//}, h.cfg.MetricIntervalCached.Value)
	//if err != nil {
	//	logger.Error(h.ctx, "cache collection is not init", zap.Error(err))
	//	return nil, fmt.Errorf("error init cache")
	//}

	//apiRouter := rt.PathPrefix("/gui/v1").Subrouter()
	//router.Use(h.JsonHeaders)

	var routes = Routes{
		// запросы (настроенные)
		Route{"Alive", "GET", "/alive", handler.Alive},
		Route{"Ping", "GET", "/ping", handler.Ping},

		// обновить роль в сессии
		Route{"AuthChangeRole", "GET", "/auth/change", handler.AuthChangeRole},
		Route{"AuthLogIn", "POST", "/auth/login", handler.AuthLogIn},
		Route{"AuthLogOut", "GET", "/auth/logout", handler.AuthLogOut},

		Route{"Cache", "GET", "/tools/cacheclear", handler.Cache},
		Route{"FileLoad", "POST", "/tools/load", handler.FileLoad},

		//Route{"Storage", "GET", "/upload/{params:.+}", handler.Storage},
		//Route{"Storage", "GET", "/assets/{params:.+}", handler.Storage},
		//Route{"Storage", "GET", "/templates/{params:.+}", handler.Storage},

		Route{"Page", "GET", "/", handler.Page},
		Route{"Page", "GET", "/{page}", handler.Page},
		Route{"Page", "POST", "/{page}", handler.Page},
		Route{"Page", "GET", "/{page}/", handler.Page},
		Route{"Page", "POST", "/{page}/", handler.Page},

		Route{"Block", "GET", "/block/{block}", handler.Block},
		Route{"Block", "POST", "/block/{block}", handler.Block},
		Route{"Block", "GET", "/block/{block}/", handler.Block},
		Route{"Block", "POST", "/block/{block}/", handler.Block},

		// Регистрация pprof-обработчиков
		Route{"pprofIndex", "GET", "/debug/pprof/", pprof.Index},
		Route{"pprofIndex", "GET", "/debug/pprof/cmdline", pprof.Cmdline},
		Route{"pprofIndex", "GET", "/debug/pprof/profile", pprof.Profile},
		Route{"pprofIndex", "GET", "/debug/pprof/symbol", pprof.Symbol},
		Route{"pprofIndex", "GET", "/debug/pprof/trace", pprof.Trace},
		Route{"pprofIndex", "GET", "/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP},
		Route{"pprofIndex", "GET", "/debug/pprof/heap", pprof.Handler("heap").ServeHTTP},
		Route{"CToolsLoadfile", "POST", "/load", handler.FileLoad},
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = h.MiddleLogger(handler, route.Name, route.Pattern)
		handler = h.XServiceKeyProcessor(handler, h.cfg)

		// проверяем адреса для исключения SSRF-уязвимостей
		// проверяем на защищенный доступ через авторизацию
		if h.cfg.SkipSecurityMiddleware != "checked" {
			handler = h.MiddleSecurity(handler, route.Name)
		}

		for _, v := range strings.Split(route.Method, ",") {
			router.
				Methods(v).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}

	router.Use(h.Recover)

	// проверяем на возможность переадресации только для HTTP запросов
	if checkHttpsOnly && h.cfg.HttpsOnly != "" {
		router.Use(h.HttpsOnly)
	}

	router.Use(h.AuthV3Middleware)

	// проверяем на защищенный доступ через авторизацию
	if h.cfg.Signin == "checked" && h.cfg.SigninUrl != "" {
		router.Use(h.AuthProcessor)
	}

	// добавление request-id в логер
	router.Use(logger.HTTPMiddleware)
	router.StrictSlash(true)

	//router.PathPrefix("/.well-known/").Handler(http.StripPrefix("/.well-known/", http.FileServer(http.Dir(h.cfg.Workingdir + "/upload"))))
	//router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(h.cfg.Workingdir + "/upload"))))
	//router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(h.cfg.Workingdir + "/templates"))))

	return router, err
}
