package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "embed"
)

//go:embed upload.html
var uploadHTML string

type fileDetails struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	const fileUploadLimit = 32 << 20 // Set max memory limit for file uploads (32MB in this case)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(fileUploadLimit)
	if err != nil {
		log.Println("Failed to parse multipart form:", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		log.Println("No files uploaded")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	fd := make([]fileDetails, len(files))

	for i, file := range files {
		f, fErr := file.Open()
		if fErr != nil {
			log.Println("Failed to open file:", err)
			continue
		}

		//nolint:errcheck // we don't care about the error here
		defer f.Close()

		content, ioErr := ioutil.ReadAll(f)
		if ioErr != nil {
			log.Println("Failed to read file content:", err)
			continue
		}

		fd[i] = fileDetails{
			Name: file.Filename,
			Size: int64(len(content)),
		}
	}

	response, err := json.Marshal(fd)
	if err != nil {
		log.Println("Failed to marshal JSON response:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response) //nolint:errcheck,gosec // we don't care about the error here
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(uploadHTML)) //nolint:errcheck,gosec // we don't care about the error here
	})

	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
