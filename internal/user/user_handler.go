package user

import (
	"fmt"
	common_csv "go_project_structure/common_pkg/csv"
	"go_project_structure/common_pkg/json"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/dto"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	UserService UserService
}

func NewUserController(_userService UserService) *UserController {
	return &UserController{
		UserService: _userService,
	}
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {

	RequestPayload := r.Context().Value("registration_payload").(dto.RegisterUserRequest)
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	user := &models.User{
		Name:     RequestPayload.Name,
		Email:    RequestPayload.Email,
		Password: RequestPayload.Password,
	}

	message, err := uc.UserService.CreateUser(user)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User registration failed.", err)
		return
	}
	responsePayload := dto.RegisterUserResponse{
		Name:  RequestPayload.Name,
		Email: RequestPayload.Email,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, responsePayload)
}

func (uc *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	var requestPayload = dto.LoginUserRequest{}
	err := json.ReadJsonBody(r, &requestPayload)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	token, err := uc.UserService.LoginUser(requestPayload.Email, requestPayload.Password)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Login failed", err)
		return
	}
	responsePayload := dto.LoginUserResponse{
		Token: token,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, "Login successful", responsePayload)
}

func (uc *UserController) GetUserById(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	userId := chi.URLParam(r, "id")

	if userId == "" {
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user id", nil)
		return
	}

	user, err := uc.UserService.GetUserById(userId)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Get user by id end point",
		"data":    user,
		"error":   nil,
	}
	json.WriteJSONResponse(w, http.StatusOK, response)

}

func (uc *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	users, err := uc.UserService.GetAllUsers()
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": "Get all users end point",
		"data":    users,
		"error":   nil,
	}
	json.WriteJSONResponse(w, http.StatusOK, response)
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	userId := chi.URLParam(r, "id")

	requestPayload := r.Context().Value("update_payload").(dto.UpdateUserRequest)

	message, err := uc.UserService.UpdateUser(userId, requestPayload.Name, requestPayload.Email)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User update failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    nil,
		"error":   nil,
	}
	json.WriteJSONResponse(w, http.StatusOK, response)
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	userId := chi.URLParam(r, "id")

	message, err := uc.UserService.DeleteUser(userId)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User delete failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    nil,
		"error":   nil,
	}
	json.WriteJSONResponse(w, http.StatusOK, response)
}

func (uc *UserController) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value("requestId")
	fmt.Println("request id: ", reqId)

	users, err := uc.UserService.GetAllUsers()
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	var userCSV []dto.UserCSV
	for _, user := range users {
		userCSV = append(userCSV, dto.UserCSV{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	fileName, err := common_csv.ExportToCSV("users", userCSV)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "CSV export failed.", err)
		return
	}

	downloadUrl := fmt.Sprintf(
		"http://localhost:3000/profile/download?file=%s",
		fileName,
	)

	response := map[string]interface{}{
		"success": true,
		"message": "Get all users end point",
		"data":    downloadUrl,
		"error":   nil,
	}
	json.WriteJSONResponse(w, http.StatusOK, response)

}

func (uc *UserController) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Missing file name", http.StatusBadRequest)
		return
	}

	// Prevent path traversal attack
	fileName = filepath.Base(fileName)

	filePath := filepath.Join("exports", fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set download headers
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "text/csv")

	// Serve file
	http.ServeFile(w, r, filePath)
}

/*
Controller
  → Read entire CSV into memory ([][]string) : If CSV has 1M rows → whole thing loads into memory
  → Service loops every row
  → Repository opens transaction PER USER: For 10,000 users → 10,000 transactions
	That is: Slow + Expensive + Heavy on WAL + Bad for performance
		- Each user insert:
			- Begin TX
			- Execute INSERT
			- Commit
  → Insert 1 row
  → Commit
*/
func (uc *UserController) UploadUserCSV(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Content-Type:", r.Header.Get("Content-Type"))

	if !strings.Contains(r.Header.Get("content-type"), "multipart/form-data") {
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", fmt.Errorf("Content-Type must be multipart/form-data"))
		return
	}

	records, err := common_csv.UploadUserCSV(r)
	if err != nil {
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "CSV upload failed.", err)
		return
	}

	message, err := uc.UserService.CreateUserViaTnx(records)

	json.WriteJsonSuccessResponse(w, http.StatusCreated, message, nil)

}
