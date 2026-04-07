package swagger

// ErrBody đáp lỗi JSON chuẩn cho Swagger.
type ErrBody struct {
	Error string `json:"error"`
}

// HealthOK đáp health check.
type HealthOK struct {
	Status string `json:"status" example:"ok"`
}
