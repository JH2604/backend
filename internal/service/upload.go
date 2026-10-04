package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"gin-demo/internal/model"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}
var ErrPhotoTooLarge = errors.New("图片超过大小限制")
var ErrPhotoType = errors.New("只支持 JPEG / PNG /WEBP格式")
var ErrUsageInvalid = errors.New("usage不在表里")

func SavePhoto(fh *multipart.FileHeader, usage string) (*model.PhotoSize, error) {
	random := make([]byte, 8)
	_, err := rand.Read(random)
	if err != nil {
		return &model.PhotoSize{}, err
	}
	photo := strconv.FormatInt(time.Now().Unix(), 10) + "_" + hex.EncodeToString(random) + filepath.Ext(fh.Filename)

	var limits = map[string]int64{
		"avatar": 2 << 20,
		"post":   5 << 20,
	}
	limit, ok := limits[usage]
	if !ok {
		return &model.PhotoSize{}, ErrUsageInvalid

	}
	if fh.Size > limit {
		return &model.PhotoSize{}, ErrPhotoTooLarge
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return &model.PhotoSize{}, err
	}
	check, err := fh.Open()
	if err != nil {
		return &model.PhotoSize{}, err
	}
	defer check.Close()

	head := make([]byte, 512)
	n, err := check.Read(head)
	if err != nil && err != io.EOF {
		return &model.PhotoSize{}, err
	}

	mimeType := http.DetectContentType(head[:n])

	if !allowedTypes[mimeType] {
		return &model.PhotoSize{}, ErrPhotoType
	}
	_, err = check.Seek(0, io.SeekStart)
	if err != nil {
		return &model.PhotoSize{}, err
	}
	config, _, err := image.DecodeConfig(check)
	if err != nil {
		return &model.PhotoSize{}, err
	}

	f, err := os.Create(filepath.Join("uploads", photo))
	if err != nil {
		return &model.PhotoSize{}, err
	}
	defer f.Close()
	_, err = check.Seek(0, io.SeekStart)
	if err != nil {
		return &model.PhotoSize{}, err
	}

	_, err = io.Copy(f, check)
	if err != nil {
		return &model.PhotoSize{}, err
	}

	return &model.PhotoSize{
		URL:    "/uploads/" + photo,
		Size:   fh.Size,
		Width:  config.Width,
		Height: config.Height,
	}, nil
}
