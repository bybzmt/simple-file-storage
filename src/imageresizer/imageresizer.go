package imageresizer

import (
	"image"
	"log"
	"os"
	"path"
	"xx/base"

	"github.com/disintegration/imaging"
	"willnorris.com/go/gifresize"

	"github.com/valyala/fasthttp"
)

func HttpHandler(ctx *fasthttp.RequestCtx) {
	data, err := urlDecode(string(ctx.Path()))
	if err != nil {
		if base.Debug {
			ctx.SetStatusCode(400)
			ctx.SetBodyString(err.Error())
			log.Println(err.Error())
		} else {
			ctx.NotFound()
		}
		return
	}

	op, anchor_s, format, width, height, file, err := decodePath(data)
	if err != nil {
		if base.Debug {
			ctx.SetStatusCode(400)
			ctx.SetBodyString(err.Error())
			log.Println(err.Error())
		} else {
			ctx.NotFound()
		}
		return
	}

	filename := path.Join(base.BaseDir, file)

	fh, err := os.Open(filename)
	if err != nil {
		if base.Debug {
			log.Println(err)
		}
		ctx.NotFound()
		return
	}
	defer fh.Close()

	fi, err := fh.Stat()
	if err != nil {
		ctx.NotFound()
		return
	}

	if fi.IsDir() {
		ctx.NotFound()
		return
	}

	//原图不转格式时不需要处理
	if op == op_ori && format == format_auto {
		ctx.SendFile(filename)
		return
	}

	//读取图片格式
	img_cfg, ori_format, err := image.DecodeConfig(fh)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString(err.Error())
		log.Println(err.Error())
		return
	}

	//为0时不改变宽高
	if width == 0 {
		width = img_cfg.Width
	}
	if height == 0 {
		height = img_cfg.Height
	}

	if format == format_auto {
		switch ori_format {
		case "png":
			format = format_png
		case "jpeg":
			format = format_jpeg
		case "gif":
			format = format_gif
		default:
			format = format_jpeg
		}
	}

	//裁切时定位参数
	var anchor imaging.Anchor
	switch anchor_s {
	case anchor_top_left:
		anchor = imaging.TopLeft
	case anchor_top:
		anchor = imaging.Top
	case anchor_top_right:
		anchor = imaging.TopRight
	case anchor_left:
		anchor = imaging.Left
	case anchor_center:
		anchor = imaging.Center
	case anchor_right:
		anchor = imaging.Right
	case anchor_bottom_left:
		anchor = imaging.BottomLeft
	case anchor_bottom:
		anchor = imaging.Bottom
	case anchor_bottom_right:
		anchor = imaging.BottomRight
	default:
		ctx.SetStatusCode(400)
		ctx.SetBodyString("unsupport anchor")
		log.Println(err.Error())
		return
	}

	var trans func(img image.Image) image.Image

	switch op {
	case op_ori:
		//原图不需要处理
		trans = func(img image.Image) image.Image {
			return img
		}
	case op_resize:
		trans = func(img image.Image) image.Image {
			return imaging.Resize(img, width, height, imaging.Lanczos)
		}
	case op_crop:
		trans = func(img image.Image) image.Image {
			return imaging.CropAnchor(img, width, height, anchor)
		}
	case op_fit:
		trans = func(img image.Image) image.Image {
			return imaging.Fit(img, width, height, imaging.Lanczos)
		}
	case op_fill:
		trans = func(img image.Image) image.Image {
			return imaging.Fill(img, width, height, anchor, imaging.Lanczos)
		}
	default:
		if base.Debug {
			ctx.SetStatusCode(400)
			ctx.SetBodyString("unsupport op")
			log.Println("unsupport anchor")
		} else {
			ctx.NotFound()
		}
		return
	}

	//重置指针
	fh.Seek(0, os.SEEK_SET)

	//动画
	if format == format_gif && ori_format == "gif" {
		ctx.SetContentType("image/gif")
		gifresize.Process(ctx, fh, trans)
		return
	}

	//非动画图
	img, ori_format, err := image.Decode(fh)
	if err != nil {
		if base.Debug {
			ctx.SetStatusCode(500)
			ctx.SetBodyString("image decode err: " + err.Error())
			log.Println("image decode err: " + err.Error())
		} else {
			ctx.NotFound()
		}
		return
	}
	img = trans(img)

	switch format {
	case format_png:
		ctx.SetContentType("image/png")
		imaging.Encode(ctx, img, imaging.PNG)
	case format_jpeg:
		ctx.SetContentType("image/jpeg")
		imaging.Encode(ctx, img, imaging.JPEG)
	case format_gif:
		ctx.SetContentType("image/gif")
		imaging.Encode(ctx, img, imaging.GIF)
	default:
		if base.Debug {
			ctx.SetStatusCode(400)
			ctx.SetBodyString("unsupport format")
			log.Println("unsupport format")
		} else {
			ctx.NotFound()
		}
		return
	}
}
