package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func AdminUploadVideo(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUploadVideo"}

	err := r.ParseMultipartForm(512 << 20)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidMultipartForm", logParams...))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("MissingFile", logParams...))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" && ext != ".m3u8" {
		HandleResponseError(w, r, NewBadRequestError("UnsupportedVideoType", logParams...))
		return
	}

	uploadDir := filepath.Join("uploads", "videos")
	err = os.MkdirAll(uploadDir, 0o755)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("create upload directory", err, logParams...))
		return
	}

	fileName := uuid.NewString() + ext
	filePath := filepath.Join(uploadDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("create video file", err, logParams...))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("save video file", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"url": "/uploads/videos/" + fileName})
}
