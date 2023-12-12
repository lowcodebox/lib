package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"git.lowcodeplatform.net/packages/cache"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"git.lowcodeplatform.net/fabric/app/pkg/servers/httpserver/handlers"
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
type prometheusReader struct {
	res prometheus.Gatherer
}

func (h *httpserver) NewRouter(checkHttpsOnly bool) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	handler := handlers.New(h.src, h.cfg)

	router.HandleFunc("/alive", handler.Alive).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	proxy, err := h.vfs.Proxy("/media", "/"+h.cfg.VfsBucket)
	if err != nil {
		logger.Panic(h.ctx, "unable init media proxy", zap.Error(err))
	}

	router.Handle("/media/{params:.+}", proxy).Methods(http.MethodGet)

	prometheus.MustRegister(version.NewCollector(h.cfg.Name))
	version.Version = h.serviceVersion
	version.Revision = h.hashCommit

	pr := prometheusReader{}
	err = cache.Cache().Register("prometheus", &pr, h.cfg.MetricIntervalCached.Value)
	if err != nil {
		logger.Panic(h.ctx, "cache collection is not init", zap.Error(err))
	}

	//apiRouter := rt.PathPrefix("/gui/v1").Subrouter()
	//router.Use(h.JsonHeaders)

	var routes = Routes{
		// запросы (настроенные)
		Route{"ProxyPing", "GET", "/ping", handler.Ping},

		// обновить роль в сессии
		Route{"ProxyPing", "GET", "/auth/change", handler.AuthChangeRole},

		Route{"Cache", "GET", "/tools/cacheclear", handler.Cache},

		Route{"Storage", "GET", "/upload/{params:.+}", handler.Storage},
		Route{"Storage", "GET", "/assets/{params:.+}", handler.Storage},
		Route{"Storage", "GET", "/templates/{params:.+}", handler.Storage},

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
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = h.MiddleLogger(handler, route.Name)

		for _, v := range strings.Split(route.Method, ",") {
			router.
				Methods(v).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}

	//router.Use(h.Recover)

	// проверяем на возможность переадресации только для HTTP запросов
	if checkHttpsOnly && h.cfg.HttpsOnly != "" {
		router.Use(h.HttpsOnly)
	}

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

	return router
}

func (p *prometheusReader) ReadSource() (res []byte, err error) {
	mf, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		err = fmt.Errorf("error prometheus Gather. err: %s", err)
		return
	}

	res, err = json.Marshal(mf)

	return res, nil
}
