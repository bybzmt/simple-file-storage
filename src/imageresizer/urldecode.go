package imageresizer

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"path"
	"path/filepath"
	"strings"
)

const op_ori = 1
const op_resize = 2
const op_crop = 3
const op_fit = 4
const op_fill = 5

const anchor_top_left = 1
const anchor_top = 2
const anchor_top_right = 3
const anchor_left = 4
const anchor_center = 5
const anchor_right = 6
const anchor_bottom_left = 7
const anchor_bottom = 8
const anchor_bottom_right = 9

const format_auto = 0
const format_jpeg = 1
const format_png = 2
const format_gif = 3

var SignatureKey string

func urlDecode(url string) (data []byte, err error) {
	var query, ext string
	query = url

	pn := strings.IndexByte(query, '.')
	if pn < 0 {
		query = query[1:]
	} else {
		ext = query[pn:]
		query = query[1:pn]
	}

	if len(query) < 9 {
		err = errors.New("data too short")
		return
	}

	raw, err := base64.RawURLEncoding.DecodeString(query)
	if err != nil {
		err = errors.New("base64 decode err" + err.Error())
		return
	}

	if raw[0] != 1 {
		err = errors.New("protol version err")
		return
	}

	data, err = checkSign(raw[1:], []byte(ext))
	return
}

func checkSign(raw, ext []byte) ([]byte, error) {
	if len(raw) < 2 {
		return nil, errors.New("checkSign data too shart")
	}

	if raw[0] == 0 {
		if SignatureKey == "" {
			return raw[1:], nil
		} else {
			return nil, errors.New("sign empty")
		}
	}

	sign_len := int(raw[0])

	if len(raw) < 1+sign_len {
		return nil, errors.New("sign data too shart")
	}

	mac := raw[1 : 1+sign_len]
	data := raw[1+sign_len:]

	if !CheckMAC(append(data, ext...), mac) {
		return nil, errors.New("sign not eq")
	}

	return data, nil
}

func CheckMAC(message, messageMAC []byte) bool {
	mac := hmac.New(md5.New, []byte(SignatureKey))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func decodePath(raw []byte) (op, anchor, format, width, height int, file string, err error) {
	if len(raw) < 6 {
		err = errors.New("data too short")
		return
	}

	op = int(raw[0] >> 4)
	anchor = int(raw[0] & 0xf)

	format = int(raw[1])

	width = int(raw[2])<<8 | int(raw[3])
	height = int(raw[4])<<8 | int(raw[5])

	file = path.Clean(string(raw[6:]))

	file, err = filepath.Rel("/", file)

	return
}
