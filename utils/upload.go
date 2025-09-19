package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func UploadImageToLocal(icode, path string, images []*multipart.FileHeader) ([]string, error) {

	// TEST

	fileNameList := []string{}
	// Create Path
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return nil, err
	}

	for _, fileHeader := range images {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MAX_UPLOAD_SIZE {
			return nil, errors.New(fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename))
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return nil, err
		}

		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			return nil, err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))

		fileNameList = append(fileNameList, newFileName)

		f, err := os.Create(path + newFileName)
		if err != nil {
			return nil, err
		}

		defer f.Close()

		pr := &Progress{
			TotalSize: fileHeader.Size,
		}
		_, err = io.Copy(f, io.TeeReader(file, pr))
		if err != nil {
			return nil, err
		}
	}

	return fileNameList, nil

}

const MAX_UPLOAD_SIZE = 1024 * 1024        // 1MB
const MAX_DOCUMENT_SIZE = 10 * 1024 * 1024 // 10MB for documents
// Progress is used to track the progress of a file upload.
// It implements the io.Writer interface so it can be passed
// to an io.TeeReader()
type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface.
// Instead of writing somewhere, it simply aggregates
// the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
// each time Write is called
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}

	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

// UploadDocumentsToLocal uploads document files (PDF, Excel, images) to local storage
func UploadDocumentsToLocal(path string, files []*multipart.FileHeader) ([]map[string]interface{}, error) {
	fileInfoList := []map[string]interface{}{}

	// Create Path
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Allowed file types
	allowedTypes := map[string]bool{
		"application/pdf":          true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}

	for _, fileHeader := range files {
		// Check file size
		if fileHeader.Size > MAX_DOCUMENT_SIZE {
			return nil, fmt.Errorf("file %s is too big: %d bytes. Maximum allowed size is %d bytes",
				fileHeader.Filename, fileHeader.Size, MAX_DOCUMENT_SIZE)
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Detect content type
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return nil, err
		}

		filetype := http.DetectContentType(buff)

		// Check file extension for Excel files (DetectContentType might not catch all Excel formats)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext == ".xlsx" || ext == ".xls" {
			if ext == ".xlsx" {
				filetype = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			} else {
				filetype = "application/vnd.ms-excel"
			}
		} else if ext == ".docx" {
			filetype = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		} else if ext == ".doc" {
			filetype = "application/msword"
		}

		// Validate file type
		if !allowedTypes[filetype] {
			return nil, fmt.Errorf("file type not allowed: %s. Allowed types: PDF, Excel, JPG, JPEG, PNG", filetype)
		}

		// Reset file pointer
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		// Generate unique filename
		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
		fullPath := filepath.Join(path, newFileName)
		// Ensure returned file paths use forward slashes for URL compatibility
		urlPath := filepath.ToSlash(fullPath)
		// Create the file
		f, err := os.Create(fullPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Copy file content
		_, err = io.Copy(f, file)
		if err != nil {
			return nil, err
		}

		// Add file info to list
		fileInfo := map[string]interface{}{
			"fileName":     newFileName,
			"originalName": fileHeader.Filename,
			"fileSize":     fileHeader.Size,
			"contentType":  filetype,
			"filePath":     urlPath,
		}
		fileInfoList = append(fileInfoList, fileInfo)
	}

	return fileInfoList, nil
}
