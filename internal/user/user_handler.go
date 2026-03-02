package user

import (
	"fmt"
	common_csv "go_project_structure/common_pkg/csv"
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/dto"
	"go_project_structure/internal/infrastructure/models"
	enums "go_project_structure/utils/enums"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	UserService    UserService
	userHandlerLog *logger.ScopeLogger
}

func NewUserController(_userService UserService) *UserController {
	return &UserController{
		UserService:    _userService,
		userHandlerLog: logger.Log.Scope("", "user", "user_handler"),
	}
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("RegisterUser")
	log.Infof("User registration start.")

	// request payload
	RequestPayload, ok := r.Context().Value(enums.CtxRegistrationPayload).(dto.RegisterUserRequest)
	if !ok {

		log.Errorf("Invalid request context")

		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request context", nil)
		return
	}

	user := &models.User{
		Name:     RequestPayload.Name,
		Email:    RequestPayload.Email,
		Password: RequestPayload.Password,
	}

	message, err := uc.UserService.CreateUser(r.Context(), user)
	if err != nil {

		log.WithFields(map[string]interface{}{
			"Name":  RequestPayload.Name,
			"Email": RequestPayload.Email,
		}).Errorf("User registration failed. %v", err)

		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User registration failed.", err)
		return
	}

	log.Infof("User registered successfully.")

	responsePayload := dto.RegisterUserResponse{
		Name:  RequestPayload.Name,
		Email: RequestPayload.Email,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, responsePayload)
}

func (uc *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("LoginUser")
	log.Infof("User login start.")

	var requestPayload = dto.LoginUserRequest{}
	err := json.ReadJsonBody(r, &requestPayload)
	if err != nil {
		log.Errorf("Invalid login request payload. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	token, err := uc.UserService.LoginUser(r.Context(), &requestPayload)
	if err != nil {
		log.Errorf("Login failed.")
		json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Login failed.", err)
		return
	}

	log.Infof("User login successfully.")

	responsePayload := dto.LoginUserResponse{
		Token: token,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, "user login successfully", responsePayload)
}

func (uc *UserController) GetUserById(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetUserById")
	log.Infof("Get user by id start.")

	userId := chi.URLParam(r, "id")

	if userId == "" {

		log.Errorf("user id is required")

		json.WriteJsonErrorResponse(
			w,
			http.StatusBadRequest,
			"Invalid user id",
			fmt.Errorf("user id is required"),
		)
		return
	}

	user, err := uc.UserService.GetUserById(r.Context(), userId)
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	log.Infof("User fetched successfully.")
	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get user by id end point", user)

}

func (uc *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetAllUsers")
	log.Infof("Get all users start.")

	users, err := uc.UserService.GetAllUsers(r.Context())
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	log.Infof("All users fetched successfully.")
	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get all users end point", users)
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("UpdateUser")
	log.Infof("User update start.")

	userId := chi.URLParam(r, "id")

	requestPayload, ok := r.Context().Value(enums.CtxUpdatePayload).(dto.UpdateUserRequest)
	if !ok {
		log.Errorf("Invalid update request context.")
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request context", nil)
		return
	}

	message, err := uc.UserService.UpdateUser(r.Context(), userId, &requestPayload)
	if err != nil {
		log.Errorf("User update failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User update failed.", err)
		return
	}

	log.Infof("User updated successfully.")
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("DeleteUser")
	log.Infof("User delete start.")

	userId := chi.URLParam(r, "id")

	message, err := uc.UserService.DeleteUser(r.Context(), userId)
	if err != nil {
		log.Errorf("User delete failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User delete failed.", err)
		return
	}

	log.Infof("User deleted successfully.")
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}

func (uc *UserController) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("ExportUsersCSV")
	log.Infof("CSV export start.")

	fileName, err := uc.UserService.ExportUsersAsCSV(r.Context())
	if err != nil {
		log.Errorf("CSV export failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "CSV export failed.", err)
		return
	}

	downloadUrl := fmt.Sprintf(
		"http://localhost:3000/api/v1/profile/download?file=%s",
		fileName,
	)

	log.Infof("CSV exported successfully. dowloaded url: %v", downloadUrl)
	json.WriteJsonSuccessResponse(w, http.StatusOK, "Export Users CSV end point", downloadUrl)

}

func (uc *UserController) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("DownloadFileHandler")
	log.Infof("File download start.")

	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		log.Errorf("Missing file name")
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "File name is required.", fmt.Errorf("Missing file name"))
		return
	}

	// Prevent path traversal attack
	fileName = filepath.Base(fileName)

	filePath := filepath.Join("exports", fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Errorf("File not found. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusNotFound, "File not found", err)
		return
	}

	// Set download headers
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "text/csv")

	// Serve file
	log.Infof("File downloaded successfully.")
	http.ServeFile(w, r, filePath)
}

func (uc *UserController) UploadUserCSV(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("UploadUserCSV")
	log.Infof("CSV upload start.")

	log.WithFields(map[string]interface{}{
		"Content-Type:": r.Header.Get("Content-Type"),
	}).Infof("content type of csv file")

	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		log.Errorf("Content-Type must be multipart/form-data")
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", fmt.Errorf("Content-Type must be multipart/form-data"))
		return
	}

	err := common_csv.UploadAndStreamCSV(r, 10, 10, uc.UserService.CreateUserViaTnxUsingBatchProcessing)
	if err != nil {
		log.Errorf("CSV upload failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "CSV upload failed.", err)
		return
	}

	log.Infof("CSV uploaded successfully.")
	json.WriteJsonSuccessResponse(w, http.StatusCreated, "CSV uploaded successfully.", nil)

}
