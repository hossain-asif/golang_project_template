package user

import (
	utils "go_project_structure/utils"
	"net/http"

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

	RequestPayload := r.Context().Value("registration_payload").(RegisterUserRequest)

	err := uc.UserService.CreateUser(
		RequestPayload.Name,
		RequestPayload.Email,
		RequestPayload.Password,
	)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User registration failed.", err)
		return
	}
	responsePayload := RegisterUserResponse{
		Name:  RequestPayload.Name,
		Email: RequestPayload.Email,
	}
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User registration suucessful", responsePayload)
}

func (uc *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload = LoginUserRequest{}
	err := utils.ReadJsonBody(r, &requestPayload)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	token, err := uc.UserService.LoginUser(requestPayload.Email, requestPayload.Password)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Login failed", err)
		return
	}
	responsePayload := LoginUserResponse{
		Token: token,
	}
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Login successful", responsePayload)
}

func (uc *UserController) GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	if userId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user id", nil)
		return
	}

	user, err := uc.UserService.GetUserById(userId)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Get user by id end point",
		"data":    user,
		"error":   nil,
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (uc *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := uc.UserService.GetAllUsers()
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User fetch failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": "Get all users end point",
		"data":    users,
		"error":   nil,
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	var requestPayload = UpdateUserRequest{}
	err := utils.ReadJsonBody(r, &requestPayload)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	message, err := uc.UserService.UpdateUser(userId, requestPayload.Name, requestPayload.Email)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User update failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    nil,
		"error":   nil,
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	message, err := uc.UserService.DeleteUser(userId)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "User delete failed.", err)
		return
	}
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    nil,
		"error":   nil,
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
