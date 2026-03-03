package middlewares

import (
	"context"
	"io"
	"net/http"
	"path/filepath"

	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/dto"
	enums "go_project_structure/utils/enums"
)

var userMiddlewareLogger = logger.Log.Scope("", "middleware", "user_middleware")

// middleware to validate the user register user payload
func UserRegisterRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := userMiddlewareLogger.Method("UserRegisterRequestValidator").WithContext(r.Context())

		var RequestPayload = dto.RegisterUserRequest{}
		if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
			log.Errorf("Json encoding error. %v", payloadErr)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
			return
		}
		log.Infof("user signup payload received.")

		// validate the payload
		if err := RequestPayload.Validate(); err != nil {
			log.Errorf("Validation failed. %v", err)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
			return
		}
		log.Infof("user signup payload validated.")

		// context can be used to pass the validated payload to the handler for further processing
		req_context := r.Context()                                                          // parent context -> get the context from the request
		ctx := context.WithValue(req_context, enums.CtxRegistrationPayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)                                                              // create a new request with the new context
		log.Infof("user signup context setup.")

		next.ServeHTTP(w, r)
	})
}

// middleware to validate the user update payload
func UserUpdateRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := userMiddlewareLogger.Method("UserUpdateRequestValidator").WithContext(r.Context())

		var RequestPayload = dto.UpdateUserRequest{}
		if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
			log.Errorf("Json encoding error. %v", payloadErr)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
			return
		}
		log.Infof("user update payload received.")

		// validate the payload
		if err := RequestPayload.Validate(); err != nil {
			log.Errorf("Validation failed. %v", err)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
			return
		}
		log.Infof("user update payload validated.")

		req_context := r.Context()                                                    // parent context -> get the context from the request
		ctx := context.WithValue(req_context, enums.CtxUpdatePayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)
		log.Infof("user update context setup.")

		next.ServeHTTP(w, r)
	})
}

// middleware to validate the login user payload
func UserLoginRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := userMiddlewareLogger.Method("UserLoginRequestValidator").WithContext(r.Context())

		var RequestPayload = dto.LoginUserRequest{}
		if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
			log.Errorf("Json encoding error. %v", payloadErr)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
			return
		}
		log.Infof("user login payload received.")

		// validate the payload
		if err := RequestPayload.Validate(); err != nil {
			log.Errorf("Validation failed. %v", err)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
			return
		}
		log.Infof("user login payload validated.")

		req_context := r.Context()                                                    // parent context -> get the context from the request
		ctx := context.WithValue(req_context, enums.CtxUpdatePayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)
		log.Infof("user login context setup.")

		next.ServeHTTP(w, r)
	})
}

// middleware to validate upload csv file

func UserUploadCSVRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := userMiddlewareLogger.Method("UserUploadCSVRequestValidator").WithContext(r.Context())

		// Limit upload size (example: 5MB)
		r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

		// Parse multipart form
		if err := r.ParseMultipartForm(5 << 20); err != nil {
			log.Errorf("Failed to parse multipart form: %v", err)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "File too large", err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Errorf("Failed to get file: %v", err)
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "File is required", err)
			return
		}
		defer file.Close()

		// Validate file extension
		// extension can be faked: malware.exe → renamed to data.csv
		// MIME detection reduces this risk.
		if filepath.Ext(header.Filename) != ".csv" {
			log.Errorf("Only CSV files are allowed")
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Only CSV files are allowed", err)
			return
		}

		// Validate MIME type
		// MIME type detection works using file signatures. All these signatures are found within the first few hundred bytes.
		// The Go standard library’s http.DetectContentType() function only looks at the first 512 bytes to determine the MIME type.
		// Less than 512 → might miss some signatures
		// More than 512 → unnecessary (function ignores extra)
		buffer := make([]byte, 512)
		file.Read(buffer)
		filetype := http.DetectContentType(buffer)

		if filetype != "text/plain; charset=utf-8" &&
			filetype != "application/vnd.ms-excel" &&
			filetype != "application/octet-stream" {

			log.Errorf("Invalid file type")
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid file type", err)
			return
		}

		// Reset file pointer after reading first 512 bytes
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Errorf("Cannot reset file pointer: %v", err)
			json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Cannot reset file pointer", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
