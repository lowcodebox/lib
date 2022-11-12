package httpserver

import (
	"context"
	"fmt"
	"git.lowcodeplatform.net/fabric/lib"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

func (h *httpserver) MiddleLogger(next http.Handler, name string, logger lib.Log, serviceMetrics lib.ServiceMetric) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		timeInterval := time.Since(start)
		if name != "ProxyPing"  { //&& false == true
			mes := fmt.Sprintf("Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				timeInterval)
			logger.Info(mes)
		}

		// сохраняем статистику всех запросов, в том числе и пинга (потому что этот запрос фиксируется в количестве)
		serviceMetrics.SetTimeRequest(timeInterval)
	})
}

func (h *httpserver) AuthProcessor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authKey string
		var err error

		// пропускаем пинги	и другие сервисные запросы
		if r.URL.Path == "/ping" || strings.Contains(r.URL.Path, "/templates") || strings.Contains(r.URL.Path, "/upload") {
			next.ServeHTTP(w, r)
			return
		}

		authKeyHeader := r.Header.Get("X-Auth-Key")
		if authKeyHeader != "" {
			authKey = authKeyHeader
		} else {
			authKeyCookie, err := r.Cookie("X-Auth-Key")
			if err == nil {
				authKey = authKeyCookie.Value
			}
		}

		// не передали ключ - вход не осуществлен. войди
		if strings.TrimSpace(authKey) == "" {
			if r.FormValue("ref") == "" {
				http.Redirect(w, r, h.cfg.SigninUrl+"?ref="+h.cfg.ClientPath+r.RequestURI, 302)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// валидируем токен
		status, token, refreshToken, err := h.iam.Verify(authKey)

		// пробуем обновить пришедший токен
		if !status {
			authKey, err = h.iam.Refresh(refreshToken, "", false)

			// если токен был обновлен чуть ранее, то текущий запрос надо пропустить
			// чтобы избежать повторного обновления и дать возможность завершиться отправленным
			// единовременно нескольким запросам (как правило это интервал 5-10 секунд)
			if authKey == "skip" {
				next.ServeHTTP(w, r)
				return
			}

			if err == nil && authKey != "<nil>" && authKey != "" {
				// заменяем куку у пользователя в браузере
				cookie := &http.Cookie{
					Path: "/",
					Name:   "X-Auth-Key",
					Value:  authKey,
					MaxAge: 30000,
				}

				// после обновления получаем текущий токен
				status, token, _, err = h.iam.Verify(authKey)

				// переписываем куку у клиента
				http.SetCookie(w, cookie)
			}
		}

		// выкидываем если обновление невозможно
		if !status || err != nil {
			if r.FormValue("ref") == "" {
				http.Redirect(w, r, h.cfg.SigninUrl+"?ref="+h.cfg.ClientPath+r.RequestURI, 302)
				return
			}
		}

		// добавляем значение токена в локальный реестр сесссий (ПЕРЕДЕЛАТЬ)
		if token != nil {
			var flagUpdateRevision bool	// флаг того, что надо обновить сессию в хранилище через запрос к IAM

			prof, _ := h.session.GetProfile(token.Session)
			if prof == nil {
				flagUpdateRevision = true
			} else {
				if prof.Revision != token.SessionRev {
					flagUpdateRevision = true
				}
			}

			// проверяем наличие сессии в локальном хранилище приложения
			// проверяем соответствие ревизии сессии из токена и в текущем хранилище
			if !h.session.Found(token.Session) || flagUpdateRevision {

				err = h.session.Set(token.Session)
				if err != nil {
					http.Redirect(w, r, h.cfg.Error500+"?err="+fmt.Sprint(err), 500)
					return
				}
			}
		}

		// добавили текущий валидный токен в заголовок запроса
		ctx := context.WithValue(r.Context(), "token", authKey)
		if token != nil {
			current_profile, _ := h.session.GetProfile(token.Session)
			ctx = context.WithValue(ctx, "profile", *current_profile)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *httpserver) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r *http.Request) {
			rec := recover()
			if rec != nil {
				b := string(debug.Stack())
				//fmt.Println(r.URL.String())
				h.logger.Panic(fmt.Errorf("%s", b), "Recover panic from path: ", r.URL.String(), "; form: ", r.Form)
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
