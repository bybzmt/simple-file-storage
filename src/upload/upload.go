package upload

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net"
	"os"
	"path"
	"strconv"
	"time"
	"unsafe"

	"github.com/valyala/fasthttp"
)

var Tmpdir = "./"

type OutFile struct {
	Filename string
	Name     string
	Tmpfile  string
}

type OutJson struct {
	Err   string
	Files []OutFile
}

func tmpfile() string {
	num := time.Now().Unix()
	var x [4]byte

	rand.Read(x[:])

	tmp := strconv.FormatInt(num, 10) + "_"
	tmp += strconv.FormatInt(int64(*(*uint32)(unsafe.Pointer(&x))), 10)

	return path.Join(Tmpdir, tmp)
}

func saveFile(conn net.Conn, file string, part *multipart.Part) error {
	f2, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println("tmpfile", file, err)
		return err
	}
	defer f2.Close()

	buf := make([]byte, 1024*16)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * 3))

		n, err := part.Read(buf)

		if n > 0 {
			_, err = f2.Write(buf[0:n])
			if err != nil {
				return err
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Println("multipart", err)
			}
			return nil
		}
	}
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	log.Println("-------------demo----------")
	header := ctx.Request.Header.String()
	log.Println(header)
	log.Println("-------------demo----------")

	var out OutJson

	err := os.MkdirAll(Tmpdir, 0755)
	if err != nil {
		log.Println(err)
		out.Err = err.Error()
		return
	}

	//输出结果
	defer json.NewEncoder(ctx).Encode(&out)

	//判断有没有文件上传头
	boundary := ctx.Request.Header.MultipartFormBoundary()
	if boundary == nil {
		out.Err = "not found multipart/form-data"
		return
	}

	pr, pw := io.Pipe()
	defer pw.Close()

	end := make(chan int, 1)

	go func() {
		defer func() {
			pr.Close()
			end <- 1
		}()

		conn := ctx.Conn()

		multi := multipart.NewReader(pr, string(boundary))

		for {
			conn.SetReadDeadline(time.Now().Add(time.Second * 3))

			part, err := multi.NextPart()
			if err != nil {
				if err != io.EOF {
					log.Println("multipart", err)
					pr.CloseWithError(err)
				}
				break
			}

			if part.FileName() != "" {
				tmpf := tmpfile()

				err := saveFile(conn, tmpf, part)
				if err != nil {
					log.Println(err)
					pr.CloseWithError(err)
					break
				}

				file := OutFile{
					Filename: part.FileName(),
					Name:     part.FormName(),
					Tmpfile:  tmpf,
				}
				out.Files = append(out.Files, file)
			}
		}
	}()

	//将body内容转换到读pipe
	err = ctx.Request.BodyWriteTo(pw)
	if err != nil {
		out.Err = err.Error()
		log.Println(err)
	}

	//等待读结束
	<-end
	close(end)
}
