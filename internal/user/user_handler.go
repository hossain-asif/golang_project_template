package user

import (
	"fmt"
	common_csv "go_project_structure/common_pkg/csv"
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/common_pkg/pagination/cursor_pagination"
	"go_project_structure/common_pkg/pagination/offset_pagination"
	"go_project_structure/common_pkg/pagination/seek_pagination"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/dto"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	UserService    UserService
	userHandlerLog *logger.ScopeLogger
	cache          *UserListCache
}

func NewUserController(_userService UserService) *UserController {
	return &UserController{
		UserService:    _userService,
		userHandlerLog: logger.Log.Scope("", "user", "user_handler"),
		cache:          NewUserListCache(1 * time.Minute),
	}
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("RegisterUser")

	// request payload
	var RequestPayload = dto.RegisterUserRequest{}
	if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
		log.Errorf("Json encoding error. %v", payloadErr)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
		return
	}

	// validate the payload
	if err := RequestPayload.Validate(); err != nil {
		log.Errorf("Validation failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
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

	uc.cache.Purge()

	responsePayload := dto.RegisterUserResponse{
		Name:  RequestPayload.Name,
		Email: RequestPayload.Email,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, responsePayload)
}

func (uc *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("LoginUser")

	// read json body
	var RequestPayload = dto.LoginUserRequest{}
	if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
		log.Errorf("Json encoding error. %v", payloadErr)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
		return
	}

	// validate the payload
	if err := RequestPayload.Validate(); err != nil {
		log.Errorf("Validation failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	token, err := uc.UserService.LoginUser(r.Context(), &RequestPayload)
	if err != nil {
		log.Errorf("Login failed.")
		json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Login failed.", err)
		return
	}

	responsePayload := dto.LoginUserResponse{
		Token: token,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, "user login successfully", responsePayload)
}

func (uc *UserController) GetUserById(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetUserById")

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

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get user by id end point", user)

}

func (uc *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetAllUsers")

	// Try in-memory cache first
	users, hit := uc.cache.Get()

	if !hit {
		log.Infof("Cache miss — fetching from DB.")

		var err error
		users, err = uc.UserService.GetAllUsers(r.Context())
		if err != nil {
			log.Errorf("User fetch failed. %v", err)
			json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
			return
		}

		uc.cache.Set(users)
	} else {
		log.Infof("Cache hit — serving from memory.")
	}

	/*
		HTTP cache headers (browser / CDN layer on top)
		public                 → browsers AND CDNs (Cloudflare, CloudFront) may cache
		max-age=86400          → cached response is fresh for 24 hours
		stale-while-revalidate → after 24h, serve stale instantly while refreshing in background
	*/
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=30")

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get all users end point", users)
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("UpdateUser")

	// read json body for update request
	var RequestPayload = dto.UpdateUserRequest{}
	if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
		log.Errorf("Json encoding error. %v", payloadErr)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
		return
	}

	// validate the payload
	if err := RequestPayload.Validate(); err != nil {
		log.Errorf("Validation failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// extract url param
	userId := chi.URLParam(r, "id")

	message, err := uc.UserService.UpdateUser(r.Context(), userId, &RequestPayload)
	if err != nil {
		log.Errorf("User update failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User update failed.", err)
		return
	}

	uc.cache.Purge()

	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("DeleteUser")

	userId := chi.URLParam(r, "id")

	message, err := uc.UserService.DeleteUser(r.Context(), userId)
	if err != nil {
		log.Errorf("User delete failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User delete failed.", err)
		return
	}

	uc.cache.Purge()

	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}

func (uc *UserController) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("ExportUsersCSV")

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

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Export Users CSV end point", downloadUrl)

}

func (uc *UserController) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("DownloadFileHandler")

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
	http.ServeFile(w, r, filePath)
}

func (uc *UserController) UploadUserCSV(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("UploadUserCSV")

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

	json.WriteJsonSuccessResponse(w, http.StatusCreated, "CSV uploaded successfully.", nil)

}

func (uc *UserController) GetUsersByOffsetPagination(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetUsersByOffsetPagination")

	// Parse & validate pagination params
	paginationParams := offset_pagination.ParseParams(r)

	// Extract any additional filters
	// _ := r.URL.Query().Get("name")

	// call service layer
	users, totalUsers, err := uc.UserService.GetUsersByOffsetPagination(r.Context(), paginationParams)
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Users fetch failed.", err)
		return
	}

	// Build paginated response
	resp := offset_pagination.NewResponse(users, paginationParams, totalUsers)

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get users by pagination end point", resp)
}

func (uc *UserController) GetUsersByCursorPagination(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetUsersByCursorPagination")

	// Parse & validate pagination params
	paginationParams, err := cursor_pagination.ParseParams(r)
	if err != nil {
		log.Errorf("Invalid pagination params. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid pagination params", err)
		return
	}

	// Extract any additional filters
	// _ := r.URL.Query().Get("name")

	// call service layer
	users, err := uc.UserService.GetUsersByCursorPagination(r.Context(), paginationParams)
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Users fetch failed.", err)
		return
	}

	// Build paginated response
	hasMore := len(users) > paginationParams.Limit
	if hasMore {
		users = users[:paginationParams.Limit]
	}

	page, err := cursor_pagination.BuildPage(users, paginationParams, hasMore)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get users by pagination end point", page)
}

func (uc *UserController) GetUsersBySeekPagination(w http.ResponseWriter, r *http.Request) {
	log := uc.userHandlerLog.WithContext(r.Context()).Method("GetUsersBySeekPagination")

	req, err := seek_pagination.ParseParams(r)
	if err != nil {
		log.Errorf("Invalid pagination params. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid pagination params", err)
		return
	}

	rail, err := uc.UserService.GetUsersBySeekPagination(r.Context(), req)
	if err != nil {
		log.Errorf("failed to fetch users. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "failed to fetch users", err)
		return
	}

	// ── 4. Count total new items for the session badge ────────────────────
	var totalNew int64
	if req.Cursor != nil {
		totalNew, _ = uc.UserService.CountUsersNewSince(r.Context(), req.Cursor.AnchorCreatedAt, req.Cursor.AnchorID)
	}

	// ── 5. Build response (generic — works for any Entity) ────────────────
	resp := seek_pagination.BuildResponse(rail, req, totalNew)

	if totalNew > 0 {
		w.Header().Set("X-New-Items", strconv.FormatInt(totalNew, 10))
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get users by pagination end point", resp)
}
