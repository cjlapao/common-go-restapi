package restapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/http"

	execution_context "github.com/cjlapao/common-go-execution-context"
	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/cjlapao/common-go/helper/reflect_helper"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type HttpListenerOptions struct {
	ApiPrefix               string
	HttpPort                string
	EnableTLS               bool
	TLSPort                 string
	TLSCertificate          string
	TLSPrivateKey           string
	UseAuthBackend          bool
	MongoDbConnectionString string
	DatabaseName            string
	EnableAuthentication    bool
	LogHealthChecks         bool
	PublicRegistration      bool
}

// HttpListener HttpListener structure
type HttpListener struct {
	Router            *mux.Router
	Context           *execution_context.Context
	Logger            *log.LoggerService
	Options           *HttpListenerOptions
	Controllers       []controllers.Controller
	DefaultAdapters   []controllers.Adapter
	Servers           []*http.Server
	shutdownRequest   chan bool
	shutdownRequested uint32
}

var globalHttpListener *HttpListener

// NewHttpListener  Creates a new controller
func NewHttpListener() *HttpListener {
	if globalHttpListener != nil {
		globalHttpListener = nil
		if len(globalHttpListener.Servers) > 0 {
			globalHttpListener.shutdownRequest <- true
		}
	}

	listener := HttpListener{
		Context: execution_context.Get(),
		Router:  mux.NewRouter().StrictSlash(true),
		Servers: make([]*http.Server, 0),
	}

	listener.shutdownRequest = make(chan bool)
	listener.Logger = log.Get()

	listener.Controllers = make([]controllers.Controller, 0)
	listener.DefaultAdapters = make([]controllers.Adapter, 0)

	listener.Options = listener.getDefaultConfiguration()

	// Appending the correlationId renewal
	listener.DefaultAdapters = append(listener.DefaultAdapters, RequestIdMiddlewareAdapter())
	listener.DefaultAdapters = append(listener.DefaultAdapters, CorrelationMiddlewareAdapter(listener.Options.LogHealthChecks))

	globalHttpListener = &listener
	return globalHttpListener
}

func GetHttpListener() *HttpListener {
	if globalHttpListener != nil {
		return globalHttpListener
	}

	return NewHttpListener()
}

func (l *HttpListener) GetApiPrefix() string {
	if l.Options.ApiPrefix == "" {
		return ""
	}

	return http_helper.JoinUrl(l.Options.ApiPrefix)
}

func (l *HttpListener) AddHealthCheck() *HttpListener {

	l.AddController(l.Probe(), http_helper.JoinUrl("health", "probe"), "GET")
	return l
}

func (l *HttpListener) AddLogger() *HttpListener {
	l.DefaultAdapters = append(l.DefaultAdapters, LoggerMiddlewareAdapter(l.Options.LogHealthChecks))
	return l
}

func (l *HttpListener) AddJsonContent() *HttpListener {
	l.DefaultAdapters = append(l.DefaultAdapters, JsonContentMiddlewareAdapter())
	return l
}

func (l *HttpListener) AddDefaultHomepage() *HttpListener {
	return l
}

func (l *HttpListener) WithPublicUserRegistration() *HttpListener {
	l.Options.PublicRegistration = true
	return l
}

func (l *HttpListener) AddController(c controllers.Controller, path string, methods ...string) {
	l.Controllers = append(l.Controllers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}

	adapters := make([]controllers.Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)

	if l.GetApiPrefix() != "" {
		path = http_helper.JoinUrl(l.GetApiPrefix(), path)
	}
	subRouter.HandleFunc(path, controllers.Adapt(
		http.HandlerFunc(c),
		adapters...).ServeHTTP)
}

func (l *HttpListener) Start() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "authorization", "Authorization", "content-type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	l.Logger.Notice("Starting %v Go Rest API v%v", l.Context.Services.Version.Name, l.Context.Services.Version.String())

	done := make(chan bool)

	l.Router.HandleFunc(l.GetApiPrefix()+"/", defaultHomepageController)
	l.Router.HandleFunc(l.GetApiPrefix()+"/shutdown", globalHttpListener.ShutdownHandler)

	// Creating and starting the http server
	srv := &http.Server{
		Addr:    ":" + l.Options.HttpPort,
		Handler: handlers.CORS(originsOk, headersOk, methodsOk)(l.Router),
	}

	l.Servers = append(l.Servers, srv)

	go func() {
		l.Logger.Info("Api listening on http://::" + l.Options.HttpPort + l.GetApiPrefix())
		l.Logger.Success("Finished Initiating http server")
		if err := srv.ListenAndServe(); err != nil {
			if !strings.Contains(err.Error(), "http: Server closed") {
				l.Logger.Error("There was an error shutting down the http server: %v", err.Error())
			}
		}
		done <- true
	}()

	if l.Options.EnableTLS {
		cert, err := tls.X509KeyPair([]byte(l.Options.TLSCertificate), []byte(l.Options.TLSPrivateKey))
		if err == nil {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			sslSrv := &http.Server{
				Addr:      ":" + l.Options.TLSPort,
				TLSConfig: tlsConfig,
				Handler:   l.Router,
			}

			l.Servers = append(l.Servers, sslSrv)

			go func() {
				l.Logger.Info("Api listening on https://::" + l.Options.TLSPort + l.GetApiPrefix())
				l.Logger.Success("Finished Initiating https server")
				if err := sslSrv.ListenAndServeTLS("", ""); err != nil {
					if !strings.Contains(err.Error(), "http: Server closed") {
						l.Logger.Error("There was an error shutting down the https server: %v", err.Error())
					}
				}
				done <- true
			}()
		} else {
			l.Logger.Error("There was an error reading the certificates to enable HTTPS")
		}
	}

	l.WaitAndShutdown()
	<-done

	l.Logger.Info("Server shut down successfully...")
}

func (l *HttpListener) WaitAndShutdown() {
	irqSign := make(chan os.Signal, 1)
	signal.Notify(irqSign, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-irqSign:
		l.Logger.Info("Server shutdown requested (signal: %v)", sig.String())
	case sig := <-l.shutdownRequest:
		l.Logger.Info("Server shutdown requested (/shutdown: %v)", fmt.Sprintf("%v", sig))
	}

	l.Logger.Info("Stopping the server...")

	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Create shutdown context with 10 second timeout
	for _, s := range l.Servers {
		err := s.Shutdown(ctx)
		if err != nil {
			l.Logger.Error("Shutdown request error: %v", err.Error())
		}
	}
}

// region Private Methods
func (l *HttpListener) getDefaultConfiguration() *HttpListenerOptions {
	options := HttpListenerOptions{
		HttpPort:                l.Context.Configuration.GetString("HTTP_PORT"),
		EnableTLS:               l.Context.Configuration.GetBool("ENABLE_TLS"),
		TLSPort:                 l.Context.Configuration.GetString("TLS_PORT"),
		TLSCertificate:          l.Context.Configuration.GetBase64("TLS_CERTIFICATE"),
		TLSPrivateKey:           l.Context.Configuration.GetBase64("TLS_PRIVATE_KEY"),
		DatabaseName:            l.Context.Configuration.GetString("MONGODB_DATABASENAME"),
		MongoDbConnectionString: l.Context.Configuration.GetBase64("MONGODB_CONNECTION_STRING"),
		LogHealthChecks:         false,
	}

	if reflect_helper.IsNilOrEmpty(options.HttpPort) {
		options.HttpPort = "5000"
	}

	if reflect_helper.IsNilOrEmpty(options.TLSPort) {
		options.TLSPort = "5001"
	}

	if reflect_helper.IsNilOrEmpty(options.DatabaseName) {
		options.DatabaseName = "users"
	}

	apiPrefix := http_helper.JoinUrl(l.Context.Configuration.GetString("API_PREFIX"))

	options.ApiPrefix = apiPrefix

	l.Options = &options

	return l.Options
}

func defaultHomepageController(w http.ResponseWriter, r *http.Request) {
	response := DefaultHomepage{
		CorrelationID: globalHttpListener.Context.CorrelationId,
		Timestamp:     fmt.Sprint(time.Now().Format(time.RFC850)),
	}

	json.NewEncoder(w).Encode(response)
}

//endregion
