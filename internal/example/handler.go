package example

import (
	"go_project_structure/common_pkg/json"
	"go_project_structure/common_pkg/logger"
	"net/http"
)

type Handler struct {
	service Service
	Log     *logger.ScopeLogger
}

func NewHandler(_service Service) *Handler {
	return &Handler{
		service: _service,
		Log:     logger.Log.Scope("", "example", "example_handler"),
	}
}

func (uc *Handler) Get(w http.ResponseWriter, r *http.Request) {
	log := uc.Log.WithContext(r.Context()).Method("Get")

	err := uc.service.Get(r.Context())
	if err != nil {
		log.Errorf("Example fetch failed. %v", err)
		json.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Example fetch failed.", err)
		return
	}

	json.WriteJsonSuccessResponse(w, http.StatusOK, "Get all examples end point", nil)
}
