package service

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func SavePhoto(fh *multipart.FileHeader) (string, error) {
	random := make([]byte, 8)
	_, err := rand.Read(random)
	if err != nil {
		return "", err
	}
	photo := strconv.FormatInt(time.Now().Unix(), 10) + "_" + hex.EncodeToString(random) + filepath.Ext(fh.Filename)
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
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
