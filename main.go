package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func reverseProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid proxy target: %v", err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `multi-proxy - A simple reverse proxy for multiple services

Usage:
  multi-proxy -r '/route1,/route2' -p 8001,8002 -sp 8080 [--debug]

Options:
`)
	flag.PrintDefaults()
}

func startDebugServer(port string, label string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("DEBUG: You hit %s at port %s\n", label, port)
		w.Write([]byte(msg))
	})
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	go func() {
		log.Printf("Starting DEBUG backend server for %s on port %s", label, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("DEBUG server error for %s: %v", label, err)
		}
	}()
}

func logRequest(route string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Printf("[ACCESS] %s %s %s %s -> %s", ip, r.Method, r.URL.Path, route, time.Since(start))
		handler.ServeHTTP(w, r)
	})
}


func main() {
	flag.Usage = usage

	routesFlag := flag.String("r", "", "Comma-separated list of route paths (e.g., /api,/auth)")
	portsFlag := flag.String("p", "", "Comma-separated list of backend ports (e.g., 8001,8002)")
	startPort := flag.String("sp", "8080", "Port for the proxy server to listen on")
	help := flag.Bool("h", false, "Show help message")
	debug := flag.Bool("debug", false, "Start dummy backend servers on given ports")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	routes := strings.Split(*routesFlag, ",")
	ports := strings.Split(*portsFlag, ",")

	if len(*routesFlag) == 0 || len(*portsFlag) == 0 || len(routes) != len(ports) {
		fmt.Println("Error: number of routes and ports must match and not be empty.")
		flag.Usage()
		os.Exit(1)
	}

	mux := http.NewServeMux()

	for i := range routes {
		route := strings.TrimSpace(routes[i])
		port := strings.TrimSpace(ports[i])
		target := fmt.Sprintf("http://localhost:%s", port)

		if *debug {
			label := fmt.Sprintf("Service %d (%s)", i+1, route)
			startDebugServer(port, label)
			time.Sleep(100 * time.Millisecond)
		}

		proxy := reverseProxy(target)
		handlerWithLogging := logRequest(route, http.StripPrefix(route, proxy))
		mux.Handle(route+"/", handlerWithLogging)
		log.Printf("Routing %s -> %s", route, target)
	}

	addr := fmt.Sprintf(":%s", *startPort)
	log.Printf("Proxy server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
