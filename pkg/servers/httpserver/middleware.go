package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

const headerReferer = "Referer"

const errorReferer = "421 Misdirected Request"

func (h *httpserver) MiddleLogger(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		timeInterval := time.Since(start)
		if name != "ProxyPing" { //&& false == true
			mes := fmt.Sprintf("Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				timeInterval)
			logger.Info(r.Context(), mes,
				zap.Float64("timing", timeInterval.Seconds()),
			)
		}

		// сохраняем статистику всех запросов, в том числе и пинга (потому что этот запрос фиксируется в количестве)
		//serviceMetrics.SetTimeRequest(timeInterval)
	})
}

// MiddleSecurity проверяем адреса для исключения SSRF-уязвимостей
func (h *httpserver) MiddleSecurity(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name != "ProxyPing" && name != "Metrics" { //&& false == true
			headerRef := r.Header.Get(headerReferer)
			if headerRef == "" {
				next.ServeHTTP(w, r.WithContext(r.Context()))
				return
			}
			if !strings.Contains(headerRef, r.Host) {
				http.Redirect(w, r, h.cfg.SigninUrl+"?error="+errorReferer, 302)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

// AuthProcessor
// проверяем на наличие токена и если есть то обогащаем контекст (если не валидный - смотрим на разрешенную страницу)
// если нет, то проверяем на разрешенные страницы
func (h *httpserver) AuthProcessor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authKey string
		var err error
		var flagPublicPages, flagPublicRoutes bool
		var currentProfile *models.ProfileData
		var skipRedirect bool
		var action = ""
		dps := h.src.GetDynamicParams()
		refURL := h.cfg.ClientPath + r.RequestURI

		if strings.Contains(r.RequestURI, "assets") || strings.Contains(r.RequestURI, "templates") {
			next.ServeHTTP(w, r)
			return
		}

		err = r.ParseForm()
		if err != nil {
			err = fmt.Errorf("error parse form for url: %s", r.URL)
			return
		}

		defer func() {
			rnd := lib.UUID()
			if action == "exit" {
				return
			}

			if authKey == "skip" {
				logger.Error(r.Context(), "auth skip after refresh", zap.String("authKey", fmt.Sprintf("%+v", authKey)), zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			// пропускаем разрешенные страницы/пути
			if flagPublicPages || flagPublicRoutes || strings.Contains(refURL, h.cfg.SigninUrl) {
				//logger.Info(r.Context(), "auth skip after public pages/block",
				//	zap.String("URL", fmt.Sprintf("%+v", r.RequestURI)),
				//	zap.String("flagPublicPages", fmt.Sprintf("%+v", flagPublicPages)),
				//	zap.String("flagPublicRoutes", fmt.Sprintf("%+v", flagPublicRoutes)),
				//	zap.String("rnd", rnd),
				//	zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if action == "redirect302" || !skipRedirect {
				logger.Error(r.Context(), "redirect302",
					zap.String("URL", fmt.Sprintf("%+v", r.RequestURI)),
					zap.String("dps", fmt.Sprintf("%+v", dps)),
					zap.String("flagPublicPages", fmt.Sprintf("%+v", flagPublicPages)),
					zap.String("flagPublicRoutes", fmt.Sprintf("%+v", flagPublicRoutes)),
					zap.String("rnd", rnd),
					zap.Error(err))

				http.Redirect(w, r, h.cfg.SigninUrl+"?ref="+refURL, 302)
			}

			if action == "redirect500" {
				http.Redirect(w, r, h.cfg.Error500+"?err="+fmt.Sprint(err), 500)
			}

			next.ServeHTTP(w, r)
		}()

		// берем токен (всегда, даже если публичная страница)
		authKeyHeader := r.Header.Get("X-Auth-Key")
		if authKeyHeader != "" {
			authKey = authKeyHeader
		} else {
			authKeyCookie, err := r.Cookie("X-Auth-Key")
			if err == nil {
				authKey = authKeyCookie.Value
			}
		}

		// условия пропуска страницы (публичная)
		for k, _ := range r.Form {
			if dps.PublicPages[k] {
				flagPublicPages = true
				break
			}
		}
		// возможно передача параметров была через /
		if !flagPublicPages {
			p := strings.Split(r.RequestURI, "?")[0] // чистим от параметров в урле
			for _, k := range strings.Split(p, "/") {
				if dps.PublicPages[k] {
					flagPublicPages = true
					break
				}
			}
			// фикс для открытой корневой страницы
			if dps.PublicPages[p] {
				flagPublicPages = true
			}
		}
		// обращение к публичному урлу
		if !flagPublicPages {
			for k, _ := range dps.PublicRoutes {
				if strings.Contains(r.URL.Path, "/"+k) {
					flagPublicRoutes = true
					break
				}
			}
		}

		// не передали ключ - вход не осуществлен. войди
		if strings.TrimSpace(authKey) != "" {

			// валидируем токен
			status, token, refreshToken, err := h.iam.Verify(h.ctx, authKey)
			logger.Info(r.Context(), "middleware iam verify before refresh",
				zap.String("token", fmt.Sprintf("%+v", token)),
				zap.String("auth key", authKey),
				zap.Bool("status", status))

			// пробуем обновить пришедший токен
			if !status {
				authKey, err = h.iam.Refresh(h.ctx, refreshToken, "", false)
				if err != nil {
					logger.Error(r.Context(), "middleware iam refresh", zap.Error(err), zap.String("refresh token", refreshToken))
				}

				// если токен был обновлен чуть ранее, то текущий запрос надо пропустить
				// чтобы избежать повторного обновления и дать возможность завершиться отправленным
				// единовременно нескольким запросам (как правило это интервал 5-10 секунд)
				if authKey == "skip" {
					return
				}

				// выкидываем если обновление невозможно
				if err != nil {
					action = "redirect302"
					return
				}

				if err == nil && authKey != "<nil>" && authKey != "" {
					// заменяем куку у пользователя в браузере
					cookie := &http.Cookie{
						Path:     "/",
						Name:     "X-Auth-Key",
						Value:    authKey,
						MaxAge:   30000,
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
					}

					// после обновления получаем текущий токен
					status, token, _, err = h.iam.Verify(h.ctx, authKey)
					if err != nil {
						logger.Error(r.Context(), "middleware iam verify after refresh", zap.Error(err), zap.String("authKey", authKey))
					} else {
						logger.Info(r.Context(), "middleware iam verify after refresh",
							zap.String("token", fmt.Sprintf("%+v", token)),
							zap.Bool("status", status))
					}

					// переписываем куку у клиента
					http.SetCookie(w, cookie)
				}
			}

			// выкидываем если обновление невозможно
			if !status || err != nil {
				if r.FormValue("ref") == "" {
					action = "redirect302"
					return
				}
			}

			logger.Info(r.Context(), "token before set to session", zap.String("token", fmt.Sprintf("%+v", token)))
			// добавляем значение токена в локальный реестр сесссий (ПЕРЕДЕЛАТЬ)
			if token != nil {
				var flagUpdateRevision bool // флаг того, что надо обновить сессию в хранилище через запрос к IAM

				prof, err := h.session.GetProfile(token.Session)
				if err != nil {
					logger.Error(r.Context(), "middleware session GetProfile", zap.Error(err), zap.String("session", token.Session))
				}

				if prof == nil {
					flagUpdateRevision = true
					//} else {
					//if prof.Revision != token.SessionRev {
					//	flagUpdateRevision = true
					//}
				}

				// проверяем наличие сессии в локальном хранилище приложения
				// проверяем соответствие ревизии сессии из токена и в текущем хранилище
				if !h.session.Found(token.Session) || flagUpdateRevision {
					logger.Info(r.Context(), "middleware session set", zap.String("token session", token.Session))
					err = h.session.Set(token.Session)
					if err != nil {
						action = "redirect500"
						return
					}
				}
			}

			// добавили текущий валидный токен в заголовок запроса
			ctx := context.WithValue(r.Context(), "token", authKey)
			logger.Info(r.Context(), "middleware context", zap.String("new token", authKey))
			if token != nil {
				currentProfile, err = h.session.GetProfile(token.Session)
				if err != nil {
					logger.Error(r.Context(), "auth error", zap.String("currentProfile", fmt.Sprintf("%+v", currentProfile)), zap.Error(err))
				}
				ctx = context.WithValue(ctx, "profile", *currentProfile)
				logger.Info(r.Context(), "middleware context", zap.String("new profile", fmt.Sprintf("%+v", currentProfile)))
			}

			if currentProfile.Uid == "" {
				logger.Info(r.Context(), "auth false",
					zap.String("authKey", authKey),
					zap.String("currentProfile", fmt.Sprintf("%+v", currentProfile)))
			}
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			action = "exit"

			return
		}

		// отдаем ответ в зависимости от состояний
		if err != nil {
			action = "redirect500"
		}
	})
}

func (h *httpserver) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r *http.Request) {
			rec := recover()
			if rec != nil {
				b := string(debug.Stack())
				logger.Panic(h.ctx, fmt.Sprintf("Recover panic from path: %s, form: %+v", r.URL.String(), r.Form), zap.String("debug stack", b))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(r)
		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) JsonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) HttpsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// remove/add not default ports from req.Host
		target := "https://" + req.Host + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		// see comments below and consider the codes 308, 302, or 301
		http.Redirect(w, req, target, http.StatusTemporaryRedirect)
	})
}
