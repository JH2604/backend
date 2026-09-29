package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const maxPhotoSize = 5 << 20

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}
var ErrPhotoTooLarge = errors.New("图片超过 5MB")
var ErrPhotoType = errors.New("只支持 JPEG / PNG 格式")

func SavePhoto(fh *multipart.FileHeader) (string, error) {
	random := make([]byte, 8)
	_, err := rand.Read(random)
	if err != nil {
		return "", err
	}
	photo := strconv.FormatInt(time.Now().Unix(), 10) + "_" + hex.EncodeToString(random) + filepath.Ext(fh.Filename)

	if fh.Size > maxPhotoSize {
		return "", ErrPhotoTooLarge
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
	}
	check, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer check.Close()

	head := make([]byte, 512)
	n, err := check.Read(head)
	if err != nil && err != io.EOF {
		return "", err
	}

	mimeType := http.DetectContentType(head[:n])

	if !allowedTypes[mimeType] {
		return "", ErrPhotoType
	}

	opened, err := fh.Open()

	if err != nil {
		return "", err
	}

	defer opened.Close()
	f, err := os.Create(filepath.Join("uploads", photo))
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, opened)
	if err != nil {
		return "", err
	}

	return "/uploads/" + photo, nil
}
