package routing

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var Shared = NewSessionMux2("")

func HFilter(pattern string, h Handler) {
	Shared.HFilter(pattern, h)
}
func HFilterFunc(pattern string, h HandleFunc) {
	Shared.HFilterFunc(pattern, h)
}
func H(pattern string, h Handler) {
	Shared.H(pattern, h)
}
func HFunc(pattern string, h HandleFunc) {
	Shared.HFunc(pattern, h)
}

var Server *http.Server
var Listner net.Listener

func ListenAndServe(addr string) (err error) {
	Server = &http.Server{Handler: Shared}
	if strings.HasPrefix(addr, "/") {
		addrs := strings.SplitN(addr, ",", 2)
		Server.Addr = addr
		Listner, err = net.Listen("unix", addrs[0])
		if err != nil {
			return
		}
		defer Listner.Close()
		var mod uint64
		mod, err = strconv.ParseUint(addrs[1], 8, 32)
		if err != nil {
			return
		}
		err = os.Chmod(addrs[0], os.FileMode(mod))
		if err != nil {
			return
		}
	} else {
		Server.Addr = addr
		Listner, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}
		Listner = &tcpKeepAliveListener{TCPListener: Listner.(*net.TCPListener)}
		defer Listner.Close()
	}
	return Server.Serve(Listner)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func HandleSignal() error {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	return Listner.Close()
}
