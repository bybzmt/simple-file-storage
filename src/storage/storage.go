package storage

import (
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"os"
	"path"
	"xx/base"
)

var fs *LocalFs

var allowIPNet *net.IPNet
var banIPNet *net.IPNet

var Prefix string

func SetInit(local_ip, remote_ip string) {
	fs = &LocalFs{RootPath: base.BaseDir}

	var err error
	if local_ip != "" {
		_, allowIPNet, err = net.ParseCIDR(local_ip)
		if err != nil {
			log.Fatalln("localIP:", local_ip, err)
		}
	}

	if remote_ip != "" {
		_, banIPNet, err = net.ParseCIDR(remote_ip)
		if err != nil {
			log.Fatalln("remoteIP:", remote_ip, err)
		}
	}
}

func chackIP(ctx *fasthttp.RequestCtx) bool {
	if banIPNet != nil {
		if banIPNet.Contains(ctx.RemoteIP()) {
			return false
		}
	}

	if allowIPNet != nil {
		if !allowIPNet.Contains(ctx.RemoteIP()) {
			return false
		}
	}

	return true
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	if !chackIP(ctx) {
		ctx.SetStatusCode(403)
		ctx.SetBodyString("No authority")
		if base.Debug {
			log.Println("403 No authority", ctx.String())
		}
		return
	}

	switch string(ctx.Method()) {
	case "GET":
		fallthrough
	case "HEAD":
		sendFile(ctx)
	case "PUT":
		saveFile(ctx)
	case "DELETE":
		deleteFile(ctx)
	default:
		ctx.SetStatusCode(405)
		ctx.SetBodyString("405 Method Not Allowed")
		if base.Debug {
			log.Println("405 Method Not Allowed", ctx.String())
		}
	}
}

//读取文件
func sendFile(ctx *fasthttp.RequestCtx) {
	file := base.SafeFile(string(ctx.Path()))
	if file == "" {
		ctx.NotFound()
	}

	f, err := fs.Open(file)
	if err != nil {
		ctx.NotFound()
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		ctx.NotFound()
		return
	}

	if d.IsDir() {
		ctx.NotFound()
		return
	}

	ctx.SendFile(file)
}

//保存文件
func saveFile(ctx *fasthttp.RequestCtx) {
	file := base.SafeFile(string(ctx.Path()))
	if file == "" {
		ctx.NotFound()
	}

	dir := path.Dir(file)
	if dir != "" {
		fs.MkdirAll(dir, os.ModePerm)
	}

	f, err := fs.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0777)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error: " + err.Error())

		if base.Debug {
			log.Println("saveFile", err.Error(), ctx.String())
		}
		return
	}
	defer f.Close()

	err = ctx.Request.BodyWriteTo(f)

	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error: " + err.Error())

		if base.Debug {
			log.Println("saveFile", err.Error(), ctx.String())
		}
		return
	}

	ctx.Write([]byte("Success"))
}

//删除文件
func deleteFile(ctx *fasthttp.RequestCtx) {
	file := base.SafeFile(string(ctx.Path()))
	if file == "" {
		ctx.NotFound()
	}

	err := os.Remove(file)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error: " + err.Error())

		if base.Debug {
			log.Println("Delete Fail ", err.Error(), ctx.String())
		}

		return
	}

	ctx.Write([]byte("Success"))
}
