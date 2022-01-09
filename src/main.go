package main

import (
	"flag"
	"log"
	//"net"
	"path"
	"runtime"
	"time"
	"xx/base"
	"xx/imageresizer"
	"xx/locker"
	"xx/storage"
	"xx/upload"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var basedir = flag.String("dir", "./", "files base dir")
var tmpdir = flag.String("tmpDir", "tmp", "upload tmp dir dir")
var addr = flag.String("addr", ":8080", "Listen addr:port")
var debug = flag.Bool("debug", true, "debug switch")
var signatureKey = flag.String("signatureKey", "", "Signature Key")
var lock_timeout = flag.Int("lockTimeout", 20, "lock timeout second")
var http_timeout = flag.Int("httpTimeout", 10, "http timeout second")
var local_ip = flag.String("localIP", "", "local ip (eg: 192.168.0.1/24)")
var remote_ip = flag.String("remoteIP", "", "remote ip (eg:10.0.0.1/32)")

func main() {
	flag.Parse()
	log.SetPrefix("simple-file-storage ")
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	runtime.GOMAXPROCS(runtime.NumCPU())

	base.BaseDir = *basedir
	base.Debug = *debug
	storage.SetInit(*local_ip, *remote_ip)
	upload.Tmpdir = path.Join(base.BaseDir, *tmpdir)
	imageresizer.SignatureKey = *signatureKey

	log.Println("BaseDir:", base.BaseDir)
	log.Println("tmpDir:", upload.Tmpdir)

	r := router.New()
	r.GET("/image/", imageresizer.HttpHandler)
	r.Handle("LOCK", "/locker", locker.HttpHandler)
	r.ANY("/file", storage.HttpHandler)
	r.ServeFiles("/static/{filepath:*}", base.BaseDir)
	r.ServeFiles("/{filepath:*}", "./assets")
	r.POST("/upload", upload.HttpHandler)

	s := fasthttp.Server{
		Handler:                      r.Handler,
		ReadTimeout:                  time.Second * time.Duration(*http_timeout),
		WriteTimeout:                 time.Second * time.Duration(*http_timeout),
		IdleTimeout:                  60 * time.Duration(*http_timeout),
		TCPKeepalive:                 true,
		DisablePreParseMultipartForm: true,
		MaxRequestBodySize:           1024 * 1024 * 1,
		StreamRequestBody:            true,
	}

	log.Println("listen on", *addr)
	log.Fatalln(s.ListenAndServe(*addr))
}
