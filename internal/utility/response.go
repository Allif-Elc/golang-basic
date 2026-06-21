package utility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

// bufferPool reuses byte buffers for JSON marshaling to reduce GC pressure.
// Initial capacity 1024 bytes matches most API response sizes.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SendSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(response); err != nil {
		w.Write([]byte(`{"status":"error","message":"internal server error","data":[]}`))
		return
	}

	w.Write(buf.Bytes())
}

func SendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Status:  "error",
		Message: message,
		Data:    []interface{}{},
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(response); err != nil {
		return
	}

	w.Write(buf.Bytes())
}
