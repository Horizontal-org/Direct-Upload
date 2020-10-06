package http

import (
	"crypto/tls"
	"github.com/horizontal-org/direct-upload/application"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

//noinspection GoNameStartsWithPackageName
type HttpServer struct {
	config      Config
	authManager *application.AuthManager
	fileStore   application.FileStore
	logger      *zap.Logger
}

type Config struct {
	Address        string
	CertFile       string
	PrivateKeyFile string
}

var fileRegexp = regexp.MustCompile("^[a-zA-Z0-9_\\-][a-zA-Z0-9_.\\-]*$")

func NewServer(cfg Config, am *application.AuthManager, fs application.FileStore, logger *zap.Logger) *HttpServer {
	return &HttpServer{
		config:      cfg,
		authManager: am,
		fileStore:   fs,
		logger:      logger,
	}
}

func (s *HttpServer) Start() {
	auth := NewBasicAuthMiddleware(s.logger, s.authManager)
	pacifier := NewPanicMiddleware(s.logger)
	logger := NewLoggerMiddleware(s.logger)

	restricted := func(h httprouter.Handle) httprouter.Handle {
		return pacifier.Handle(logger.Handle(auth.Handle(h)))
	}

	router := httprouter.New()
	router.HEAD("/:file", restricted(s.handleHead))
	router.PUT("/:file", restricted(s.handlePut))
	router.POST("/:file", restricted(s.handlePost))
	router.DELETE("/:file", restricted(s.handleDelete))

	s.logger.Sugar().Infof("Starting Tella upload server on %s", s.config.Address)

	s.logger.Fatal("Error on Tella upload server start", zap.Error(s.listen(router)))
}

func (s *HttpServer) handleHead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := ps.ByName("file")

	// validate parameters
	if !validFileName(file) {
		errorValidation(w)
		return
	}

	fileInfo, err := s.fileStore.GetFileInfo(r.Context(), file)
	if err != nil {
		errorInternal(w)
		return
	}

	w.Header().Set("content-length", strconv.FormatInt(fileInfo.Size, 10))
}

func (s *HttpServer) handlePut(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := ps.ByName("file")

	// validate parameters
	if !validFileName(file) {
		errorValidation(w)
		return
	}

	err := s.fileStore.AppendFile(r.Context(), file, r.Body)

	if err == application.ErrConflict {
		errorConflict(w)
		return
	}

	if err != nil {
		errorInternal(w)
		return
	}

	ok(w)
}

func (s *HttpServer) handlePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := ps.ByName("file")

	// validate parameters
	if !validFileName(file) {
		errorValidation(w)
		return
	}

	err := s.fileStore.CloseFile(r.Context(), file)
	if err != nil {
		errorInternal(w)
		return
	}

	ok(w)
}

func (s *HttpServer) handleDelete(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	file := ps.ByName("file")

	// validate parameters
	if !validFileName(file) {
		errorValidation(w)
		return
	}

	ok(w)
}

func (s *HttpServer) listen(router *httprouter.Router) error {
	srv := http.Server{
		Addr:              s.config.Address,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       0,
		WriteTimeout:      0,
	}

	if s.config.CertFile != "" && s.config.PrivateKeyFile != "" {
		srv.TLSConfig = &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			},
		}

		return srv.ListenAndServeTLS(s.config.CertFile, s.config.PrivateKeyFile)
	}

	return srv.ListenAndServe()
}

func validFileName(str string) bool {
	return fileRegexp.MatchString(str)
}
