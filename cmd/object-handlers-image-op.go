package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/sergeymakinen/go-bmp"
)

func (api objectAPIHandlers) processImage(r *http.Request, gr *GetObjectReader) ([]byte, error) {
	//是图片文件
	objInfo := gr.ObjInfo
	objectLowerName := strings.ToLower(objInfo.Name)
	// fmt.Println("objectLowerName:", objectLowerName, objInfo.Bucket)

	// vars := mux.Vars(r)
	// fmt.Println("vars:", vars)
	x := r.FormValue("x")
	y := r.FormValue("y")
	w := r.FormValue("w")
	h := r.FormValue("h")

	crop_x, err1 := strconv.Atoi(x)
	crop_y, err2 := strconv.Atoi(y)
	crop_w, err3 := strconv.Atoi(w)
	crop_h, err4 := strconv.Atoi(h)

	// fmt.Println("r.Form():", crop_x, crop_y, crop_w, crop_h)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, fmt.Errorf("request params error: x:%v, y:%v, w:%v, h:%v", x, y, w, h)
	}

	if crop_w <= 0 || crop_h <= 0 {
		return nil, fmt.Errorf("width=%v, height=%v", crop_w, crop_h)
	}

	// if filename == "" && (strings.HasSuffix(objectLowerName, "jpg") || strings.HasSuffix(objectLowerName, "png") || strings.HasSuffix(objectLowerName, "jpeg") || strings.HasSuffix(objectLowerName, "bmp")) {
	if !strings.HasSuffix(objectLowerName, ".jpg") && !strings.HasSuffix(objectLowerName, "png") &&
		!strings.HasSuffix(objectLowerName, "jpeg") && !strings.HasSuffix(objectLowerName, "bmp") {
		return nil, fmt.Errorf("objectLowerName just support [jpg,jpeg,png,bmp] now :%v", objectLowerName)
	}

	//存在图片缩略参数
	m, format, decodeErr := image.Decode(bufio.NewReader(gr.Reader))
	if decodeErr != nil {
		return nil, fmt.Errorf("format:%v, error:%v", format, decodeErr.Error())
	}

	image_w := m.Bounds().Size().X
	image_h := m.Bounds().Size().Y

	if crop_x < 0 {
		crop_x = 0
	}
	if crop_y < 0 {
		crop_y = 0
	}
	if crop_w >= image_w {
		crop_w = image_w - 1
	}
	if crop_h >= image_h {
		crop_h = image_h - 1
	}

	if crop_x+crop_w >= image_w {
		crop_x = image_w - crop_w
	}
	if crop_y+crop_h >= image_h {
		crop_y = image_h - crop_h
	}

	var subImage image.Image
	if m.ColorModel() == color.GrayModel {
		gray := m.(*image.Gray)
		subImage = gray.SubImage(image.Rect(crop_x, crop_y, crop_x+crop_w, crop_y+crop_h)).(*image.Gray)
	} else if m.ColorModel() == color.RGBAModel {
		rgba := m.(*image.RGBA)
		subImage = rgba.SubImage(image.Rect(crop_x, crop_y, crop_x+crop_w, crop_y+crop_h)).(*image.RGBA)
	} else {
		bw := m.(*image.Paletted)
		subImage = bw.SubImage(image.Rect(crop_x, crop_y, crop_x+crop_w, crop_y+crop_h)).(*image.Paletted)
	}

	buff := new(bytes.Buffer)
	err := jpeg.Encode(buff, subImage, &jpeg.Options{Quality: 70})
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode Error: %v", err)
	}

	return buff.Bytes(), nil
}
