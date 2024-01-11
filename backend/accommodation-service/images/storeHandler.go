// package images

// import (
// 	"Rest/storage"
// 	"io"
// 	"log"
// 	"net/http"
// )

// type KeyProduct struct{}

// type StorageHandler struct {
// 	logger *log.Logger
// 	// NoSQL: injecting file storage
// 	store *storage.FileStorage
// }

// func NewStorageHandler(l *log.Logger, s *storage.FileStorage) *StorageHandler {
// 	return &StorageHandler{l, s}
// }

// func (s *StorageHandler) CopyFileToStorage(rw http.ResponseWriter, h *http.Request) {
// 	fileName := h.FormValue("fileName")

// 	err := s.store.CopyLocalFile(fileName, fileName)

// 	if err != nil {
// 		http.Error(rw, "File storage exception", http.StatusInternalServerError)
// 		s.logger.Println("File storage exception: ", err)
// 		return
// 	}
// }

// func (s *StorageHandler) WriteFileToStorage(rw http.ResponseWriter, h *http.Request) {
// 	fileName := h.FormValue("fileName")

// 	// NoSQL TODO: expand method so that it accepts file from request
// 	fileContent := "Hola Mundo!"

// 	err := s.store.WriteFile(fileContent, fileName)

// 	if err != nil {
// 		http.Error(rw, "File storage exception", http.StatusInternalServerError)
// 		s.logger.Println("File storage exception: ", err)
// 	}
// }

// func (s *StorageHandler) ReadFileFromStorage(rw http.ResponseWriter, h *http.Request) {
// 	fileName := h.FormValue("fileName")
// 	copied := h.FormValue("isCopied")
// 	isCopied := false
// 	if copied != "" {
// 		isCopied = true
// 	}

// 	fileContent, err := s.store.ReadFile(fileName, isCopied)

// 	if err != nil {
// 		http.Error(rw, "File storage exception", http.StatusInternalServerError)
// 		s.logger.Println("File storage exception: ", err)
// 		return
// 	}

// 	// Write content to response
// 	io.WriteString(rw, fileContent)
// 	s.logger.Printf("Content of file %s: %s\n", fileName, fileContent)
// }
