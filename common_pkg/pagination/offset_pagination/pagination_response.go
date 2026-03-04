package offsetpagination

// Response is a generic paginated API response wrapper
type Response[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// NewResponse creates a paginated response with data and metadata.
func NewResponse[T any](data []T, p Params, totalItems int64) Response[T] {
	if data == nil {
		data = []T{}
	}
	return Response[T]{
		Data: data,
		Meta: NewMeta(p, totalItems),
	}
}
