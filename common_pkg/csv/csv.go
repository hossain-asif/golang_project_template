package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"go_project_structure/common_pkg/logger"
	workerpool "go_project_structure/common_pkg/worker_pool"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"
)

var csvLog = logger.Log.Scope("common_pkg", "csv", "csv")

// ExportToCSV is a generic function that exports ANY slice of structs into a CSV file.
// filePrefix determines the filename prefix (e.g., users, roles, permissions)
// data must be a slice of structs
func ExportToCSV(filePrefix string, data interface{}) (string, error) {
	log := csvLog.Method("ExportToCSV")

	// Convert the input data into a reflection value
	// This allows us to inspect it at runtime
	val := reflect.ValueOf(data)

	// Ensure that the provided data is actually a slice
	// Because CSV export expects multiple records
	if val.Kind() != reflect.Slice {
		log.Errorf("data must be a slice")
		return "", fmt.Errorf("data must be a slice")
	}

	// Prevent creating a CSV file if the slice is empty
	// Because we cannot infer struct fields without at least one element
	if val.Len() == 0 {
		log.Errorf("empty slice")
		return "", fmt.Errorf("empty slice")
	}

	// Create filename with timestamp (contains : which is invalid in Windows filenames.)
	// alternative: use time.Now().Format("20060102_150405")
	// fileName := "exports/" + filePrefix + "_" + time.Now().Format("2006-01-02 15:04:05") + ".csv"
	fileName := "exports/" + filePrefix + "_" + time.Now().Format("20060102_150405") + ".csv"

	// Create a new file in the filesystem. If file already exists, it will be overwritten.
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("failed to create file: %v", err)
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
	writeHeaderErr := writer.Write(headers)
	if writeHeaderErr != nil {
		log.Errorf("failed to write header: %v", writeHeaderErr)
		return "", writeHeaderErr
	}

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
		writeRowErr := writer.Write(row)
		if writeRowErr != nil {
			log.Errorf("failed to write row: %v", writeRowErr)
			return "", writeRowErr
		}
	}

	// Return the generated filename if everything succeeded
	log.Infof("exported to %s", fileName)
	return fileName, nil
}

func UploadAndStreamCSV(r *http.Request, batchSize int, workerCount int, process func(ctx context.Context, records [][]string) error) error {
	log := csvLog.Method("UploadAndStreamCSV").WithContext(r.Context())

	uploadedFile, _, fileErr := r.FormFile("file")
	if fileErr != nil {
		log.Errorf("File is required: %v", fileErr)
		return fmt.Errorf("File is required: %v", fileErr)
	}
	defer uploadedFile.Close()

	reader := csv.NewReader(uploadedFile)

	// Create worker pool for csv parellel processing
	pool := workerpool.NewPool(r.Context(), workerCount, process)

	records := make([][]string, 0, batchSize)

	// remove header : first row that contains column names
	csvHeader, csvHeaderErr := reader.Read()
	if csvHeaderErr != nil {
		log.Errorf("invalid CSV header: %v", csvHeaderErr)
		return fmt.Errorf("invalid CSV header: %v", csvHeaderErr)
	}

	log.Infof("csv header: %v", csvHeader)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pool.Done()
			log.Errorf("error reading CSV: %v", err)
			return fmt.Errorf("error reading CSV: %v", err)
		}
		records = append(records, record)

		if len(records) == batchSize {
			batch := make([][]string, len(records))
			copy(batch, records)
			pool.Submit(batch)
			records = records[:0]
		}
	}

	// remaining records that are less than batchSize
	if len(records) > 0 {
		pool.Submit(records)
	}
	pool.Done() // no more jobs

	// Collect results/errors
	var firstErr error
	for result := range pool.Results() {
		if result.Err != nil && firstErr == nil {
			log.Errorf("error processing CSV: %v", result.Err)
			firstErr = result.Err
		}
	}

	return firstErr
}
