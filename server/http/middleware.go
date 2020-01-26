package http

import (
	"github.com/julienschmidt/httprouter"
	"github.com/tomislavr/tus/application"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

type Middleware interface {
	Handle(h httprouter.Handle) httprouter.Handle
}

type BasicAuthMiddleware struct {
	manager *application.AuthManager
	logger  *zap.Logger
}

func NewBasicAuthMiddleware(logger *zap.Logger, manager *application.AuthManager) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		manager: manager,
		logger:  logger,
	}
}

func (m *BasicAuthMiddleware) Handle(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, password, hasAuth := r.BasicAuth()

		if hasAuth {
			if !application.ValidUsername(user) {
				m.logger.Debug("Username not valid", zap.String("username", user))
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			ok, err := m.manager.CheckPassword(user, password)
			if err != nil {
				m.logger.Error("Error while validating credentials", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if ok {
				ctx := application.NewContext(r.Context(), &application.User{Username: user})
				h(w, r.WithContext(ctx), ps)
				return
			}

			m.logger.Debug("Bad credentials", zap.String("username", user))
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

type PanicMiddleware struct {
	logger *zap.Logger
}

func NewPanicMiddleware(logger *zap.Logger) *PanicMiddleware {
	return &PanicMiddleware{
		logger: logger,
	}
}

func (m *PanicMiddleware) Handle(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		defer func() {
			if rec := recover(); rec != nil {
				m.logger.Error("PanicMiddleware",
					zap.String("method", r.Method), zap.String("url", r.URL.Path),
					zap.Any("recover", rec),
					zap.String("stack", string(debug.Stack())))
			}
		}()
		h(w, r, ps)
	}
}

type LoggerMiddleware struct {
	logger *zap.Logger
}

func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

func (m *LoggerMiddleware) Handle(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		m.logger.Info("HTTP Request", zap.String("method", r.Method), zap.String("url", r.URL.Path))
		h(w, r, ps)
	}
}
