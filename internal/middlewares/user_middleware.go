package middlewares

import (
	"context"
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/dto"
	enums "go_project_structure/utils/enums"
	"net/http"
)

var userMiddlewareLogger = logger.Log.Scope("", "middleware", "user_middleware")

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

		// validation logic for the user registration payload

		// context can be used to pass the validated payload to the handler for further processing
		req_context := r.Context()                                                          // parent context -> get the context from the request
		ctx := context.WithValue(req_context, enums.CtxRegistrationPayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)                                                              // create a new request with the new context

		next.ServeHTTP(w, r)
	})
}

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

		// validation logic for the user registration payload

		req_context := r.Context()                                                    // parent context -> get the context from the request
		ctx := context.WithValue(req_context, enums.CtxUpdatePayload, RequestPayload) // create a new context with the validated payload
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
