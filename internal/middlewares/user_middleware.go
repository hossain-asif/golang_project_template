package middlewares

import (
	"context"
	"fmt"
	"go_project_structure/internal/dto"
	"go_project_structure/common_pkg/json"
	"net/http"
)

func UserRegisterRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var RequestPayload = dto.RegisterUserRequest{}
		if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
			return
		}
		fmt.Println("register payload received.")

		// validation logic for the user registration payload

		// context can be used to pass the validated payload to the handler for further processing
		req_context := r.Context()                                                    // parent context -> get the context from the request
		ctx := context.WithValue(req_context, CtxRegistrationPayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)                                                        // create a new request with the new context

		next.ServeHTTP(w, r)
	})
}

func UserUpdateRequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var RequestPayload = dto.UpdateUserRequest{}
		if payloadErr := json.ReadJsonBody(r, &RequestPayload); payloadErr != nil {
			json.WriteJsonErrorResponse(w, http.StatusBadRequest, "Json encoding error.", payloadErr)
			return
		}
		fmt.Println("update payload received.")

		// validation logic for the user registration payload

		req_context := r.Context()                                              // parent context -> get the context from the request
		ctx := context.WithValue(req_context, CtxUpdatePayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
