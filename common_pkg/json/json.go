package json

import (
	"encoding/json"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
	w.WriteHeader(statusCode)                          // Set the HTTP status code
	encoder := json.NewEncoder(w)                      // Create a JSON encoder -> create json for the response body
	return encoder.Encode(data)                        // Encode the data to JSON and write it to the response
}

func WriteJsonSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) error {
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
		"error":   nil,
	}
	return WriteJSONResponse(w, statusCode, response)
}

func WriteJsonErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) error {
	response := map[string]interface{}{}
	response["success"] = false
	response["message"] = message
	response["data"] = nil
	response["error"] = err.Error()
	return WriteJSONResponse(w, statusCode, response)
}

func ReadJsonBody(r *http.Request, result interface{}) error {
	decoder := json.NewDecoder(r.Body) // Create a JSON decoder for the request body -> read Json from the request
	// decoder.DisallowUnknownFields()    // Disallow unknown fields in the JSON to prevent errors from unexpected data
	return decoder.Decode(result)      // Decode the JSON from the request body into the destination struct
}
