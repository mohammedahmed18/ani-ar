package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type ServerConfig struct {
	// ShowStartBanner indicates whether to show or hide the server start console message.
	ShowStartBanner bool

	// HttpAddr is the TCP address to listen for the HTTP server (eg. `127.0.0.1:80`).
	HttpAddr string

	// HttpsAddr is the TCP address to listen for the HTTPS server (eg. `127.0.0.1:443`).
	// TODO: implement https
	// HttpsAddr string

	// Optional domains list to use when issuing the TLS certificate.
	//
	// If not set, the host from the bound server address will be used.
	//
	// For convenience, for each "non-www" domain a "www" entry and
	// redirect will be automatically added.
	// CertificateDomains []string

	// AllowedOrigins is an optional list of CORS origins (default to "*").
	AllowedOrigins []string

	TimeToWaitBeforeGracefulShutdown time.Duration
}

func Serve(cfg *ServerConfig) (*http.Server, error) {
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"*"}
	}

	app := InitApp()
	InitiateRoutes(app)

	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(cfg.AllowedOrigins, ", "),
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}, ","),
	}))

	// start http server
	// ---
	mainAddr := cfg.HttpAddr
	// if cfg.HttpsAddr != "" {
	// 	mainAddr = cfg.HttpsAddr
	// }

	// var wwwRedirects []string

	// extract the host names for the certificate host policy
	// hostNames := cfg.CertificateDomains
	// if len(hostNames) == 0 {
	// 	host, _, _ := net.SplitHostPort(mainAddr)
	// 	hostNames = append(hostNames, host)
	// }
	// for _, host := range hostNames {
	// 	if strings.HasPrefix(host, "www.") {
	// 		continue // explicitly set www host
	// 	}

	// 	wwwHost := "www." + host
	// 	if !list.ExistInSlice(wwwHost, hostNames) {
	// 		hostNames = append(hostNames, wwwHost)
	// 		wwwRedirects = append(wwwRedirects, wwwHost)
	// 	}
	// }

	// implicit www->non-www redirect(s)
	// if len(wwwRedirects) > 0 {
	// 	router.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
	// 		return func(c echo.Context) error {
	// 			host := c.Request().Host

	// 			if strings.HasPrefix(host, "www.") && list.ExistInSlice(host, wwwRedirects) {
	// 				return c.Redirect(
	// 					http.StatusTemporaryRedirect,
	// 					(c.Scheme() + "://" + host[4:] + c.Request().RequestURI),
	// 				)
	// 			}

	// 			return next(c)
	// 		}
	// 	})
	// }

	// certManager := &autocert.Manager{
	// 	Prompt:     autocert.AcceptTOS,
	// 	Cache:      autocert.DirCache(filepath.Join(app.DataDir(), ".autocert_cache")),
	// 	HostPolicy: autocert.HostWhitelist(hostNames...),
	// }

	// base request context used for cancelling long running requests
	// like the SSE connections
	baseCtx, cancelBaseCtx := context.WithCancel(context.Background())
	defer cancelBaseCtx()

	server := &http.Server{
		// TLSConfig: &tls.Config{
		// 	MinVersion:     tls.VersionTLS12,
		// 	GetCertificate: certManager.GetCertificate,
		// 	NextProtos:     []string{acme.ALPNProto},
		// },
		Handler:           adaptor.FiberApp(app),
		ReadTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		// WriteTimeout: 60 * time.Second, // breaks sse!
		Addr: mainAddr,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
	}

	if cfg.ShowStartBanner {
		schema := "http"
		addr := server.Addr

		// if cfg.HttpsAddr != "" {
		// 	schema = "https"

		// 	if len(cfg.CertificateDomains) > 0 {
		// 		addr = cfg.CertificateDomains[0]
		// 	}
		// }

		date := new(strings.Builder)
		log.New(date, "", log.LstdFlags).Print()

		bold := color.New(color.Bold).Add(color.FgGreen)
		bold.Printf(
			"%s Server started at %s\n",
			strings.TrimSpace(date.String()),
			color.CyanString("%s://%s", schema, addr),
		)

		regular := color.New()
		regular.Printf("├─ REST API: %s\n", color.CyanString("%s://%s/api/", schema, addr))
		regular.Printf("├─ Anime Results API: %s\n", color.CyanString("%s://%s/api/ani-results", schema, addr))
		regular.Printf("├─ Anime Episodes API: %s\n", color.CyanString("%s://%s/api/ani-episodes", schema, addr))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		_ = <-c
		// wait for execve and other handlers up to 5 seconds before exit
		ttw := cfg.TimeToWaitBeforeGracefulShutdown // time to wait
		if ttw == 0 {
			ttw = time.Second * 5
		}
		fmt.Printf("Gracefully shutting down..., waiting %v seconds\n", ttw.Seconds())
		time.AfterFunc(ttw, func() {
			server.Shutdown(baseCtx)
		})
	}()

	// @todo consider removing the server return value because it is
	// not really useful when combined with the blocking serve calls
	// ---

	// start HTTPS server
	// if cfg.HttpsAddr != "" {
	// 	// if httpAddr is set, start an HTTP server to redirect the traffic to the HTTPS version
	// 	if cfg.HttpAddr != "" {
	// 		go http.ListenAndServe(cfg.HttpAddr, certManager.HTTPHandler(nil))
	// 	}

	// 	return server, server.ListenAndServeTLS("", "")
	// }

	// OR start HTTP server
	return server, server.ListenAndServe()
}
