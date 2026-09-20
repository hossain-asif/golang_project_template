package identity

import (
	"fmt"
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/dto/identity"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
	Log     *logger.ScopeLogger
}

func NewHandler(_service Service) *Handler {
	return &Handler{
		service: _service,
		Log:     logger.Log.Scope("", "identity", "identity_handler"),
	}
}

func (uc *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("RegisterUser")

	// request payload
	var RequestPayload = identityDTO.RegisterUserRequest{}
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

	message, err := uc.service.CreateUser(r.Context(), user)
	if err != nil {

		log.WithFields(map[string]interface{}{
			"Name":  RequestPayload.Name,
			"Email": RequestPayload.Email,
		}).Errorf("User registration failed. %v", err)

		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User registration failed.", err)
		return
	}

	// uc.cache.Purge()

	responsePayload := identityDTO.RegisterUserResponse{
		Name:  RequestPayload.Name,
		Email: RequestPayload.Email,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, message, responsePayload)
}

func (uc *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("LoginUser")

	// read json body
	var RequestPayload = identityDTO.LoginUserRequest{}
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

	token, err := uc.service.LoginUser(r.Context(), &RequestPayload)
	if err != nil {
		log.Errorf("Login failed.")
		json.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Login failed.", err)
		return
	}

	responsePayload := identityDTO.LoginUserResponse{
		Token: token,
	}
	json.WriteJsonSuccessResponse(w, http.StatusOK, "user login successfully", responsePayload)
}

func (uc *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("GetUserById")

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

	user, err := uc.service.GetUserById(r.Context(), userId)
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get user by id end point", user)

}

func (uc *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("GetAllUsers")

	users, err := uc.service.GetAllUsers(r.Context())
	if err != nil {
		log.Errorf("User fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get all users end point", users)
}

func (uc *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("UpdateUser")

	// read json body for update request
	var RequestPayload = identityDTO.UpdateUserRequest{}
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

	message, err := uc.service.UpdateUser(r.Context(), userId, &RequestPayload)
	if err != nil {
		log.Errorf("User update failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User update failed.", err)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}

func (uc *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("DeleteUser")

	userId := chi.URLParam(r, "id")

	message, err := uc.service.DeleteUser(r.Context(), userId)
	if err != nil {
		log.Errorf("User delete failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User delete failed.", err)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, message, nil)
}
