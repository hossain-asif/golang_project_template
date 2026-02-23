package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// ExportToCSV is a generic function that exports ANY slice of structs into a CSV file.
// filePrefix determines the filename prefix (e.g., users, roles, permissions)
// data must be a slice of structs
func ExportToCSV(filePrefix string, data interface{}) (string, error) {

	// Convert the input data into a reflection value
	// This allows us to inspect it at runtime
	val := reflect.ValueOf(data)

	// Ensure that the provided data is actually a slice
	// Because CSV export expects multiple records
	if val.Kind() != reflect.Slice {
		return "", fmt.Errorf("data must be a slice")
	}

	// Prevent creating a CSV file if the slice is empty
	// Because we cannot infer struct fields without at least one element
	if val.Len() == 0 {
		return "", fmt.Errorf("empty slice")
	}

	// Create filename with timestamp (contains : which is invalid in Windows filenames.)
	// alternative: use time.Now().Format("20060102_150405")
	fileName := "exports/" + filePrefix + "_" + time.Now().Format("2006-01-02 15:04:05") + ".csv"

	// Create a new file in the filesystem. If file already exists, it will be overwritten.
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	// Ensure file is properly closed after function execution
	defer file.Close()

	// Create a CSV writer that writes into the file
	writer := csv.NewWriter(file)

	// Ensure any buffered CSV data is written to file before exiting
	defer writer.Flush()

	// Get the type of the first element in the slice
	// This is used to inspect struct fields (for header generation)
	elemType := val.Index(0).Type()

	// Stores column names for CSV header
	var headers []string

	// Loop through each field of the struct
	for i := 0; i < elemType.NumField(); i++ {

		// Get metadata of the struct field (name, tag, type, etc.)
		field := elemType.Field(i)

		// Extract the "csv" struct tag value
		// Example: `csv:"email"` → returns "email"
		tag := field.Tag.Get("csv")

		// Only include fields that explicitly define a csv tag
		// This allows selective exporting of fields
		if tag != "" {
			headers = append(headers, tag)
		}
	}

	// Write the header row into the CSV file
	writer.Write(headers)

	// Loop over each element (each struct instance) in the slice
	for i := 0; i < val.Len(); i++ {

		var row []string // Represents a single CSV row

		// Get the current struct value
		elem := val.Index(i)

		// Loop over each field of the struct
		for j := 0; j < elem.NumField(); j++ {

			// Get field metadata again (needed to access tag)
			fieldType := elemType.Field(j)

			// Retrieve csv tag for this field
			tag := fieldType.Tag.Get("csv")

			// Skip fields that do not have csv tag
			// This keeps behavior consistent with header logic
			if tag == "" {
				continue
			}

			// Extract actual field value from struct instance
			fieldValue := elem.Field(j).Interface()

			// Convert any type (int, string, time, etc.) into string
			// CSV requires string representation
			row = append(row, fmt.Sprintf("%v", fieldValue))
		}

		// Write the row into CSV file
		writer.Write(row)
	}

	// Return the generated filename if everything succeeded
	return fileName, nil
}

func UploadCSV(r *http.Request) ([][]string, error) {
	err := r.ParseMultipartForm(10 << 20) // file size 10MB
	if err != nil {
		return nil, fmt.Errorf("File too large: %v", err)
	}

	uploadedFile, uploadedFileHeader, fileErr := r.FormFile("file")
	if fileErr != nil {
		return nil, fmt.Errorf("File too large: %v", fileErr)
	}
	defer uploadedFile.Close()

	// Save file in uploads directory
	filePath := filepath.Join("upload", uploadedFileHeader.Filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("File already exists: %v", err)
	}

	// Create file
	dstFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, uploadedFile)
	if err != nil {
		return nil, err
	}

	// Open file
    savedFile, fileOpenErr := os.Open(filePath)
	if fileOpenErr != nil {
		return nil, fmt.Errorf("Failed to open file %v", fileOpenErr)
	}
	defer savedFile.Close()

	// Read all records
	// ReadAll() is not recommended because
	// - It loads entire CSV into memory
	// - Bad for 1M+ rows
	// records, err := reader.ReadAll()
	// if err != nil {
	// 	return nil, err
	// }

	// remove header : first row that contains column names
	// if len(records) > 1 {
	// 	records = records[1:]
	// }

	reader := csv.NewReader(savedFile)

	records := make([][]string, 0)

	// remove header : first row that contains column names
	csvHeader, csvHeaderErr := reader.Read()
	if csvHeaderErr != nil {
		return nil, fmt.Errorf("invalid CSV header: %v", csvHeaderErr)
	}
	
	fmt.Println("csv header: ", csvHeader)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %v", err)
		}
		records = append(records, record)
	}
	return records, nil
}
