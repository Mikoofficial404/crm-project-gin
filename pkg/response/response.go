package response

import "github.com/sirupsen/logrus"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
}

type Meta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

func Success(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func Error(message string) Response {
	return Response{
		Success: false,
		Message: message,
		Data:    nil,
	}
}

func ServerError(err error, userMessage string) Response {
	logrus.WithError(err).Error(userMessage)
	return Response{
		Success: false,
		Message: userMessage,
		Data:    nil,
	}
}

func Paginated(data interface{}, total int64, page int, limit int) PaginatedResponse {
	return PaginatedResponse{
		Success: true,
		Message: "Berhasil",
		Data:    data,
		Meta: Meta{
			Total: total,
			Page:  page,
			Limit: limit,
		},
	}
}
