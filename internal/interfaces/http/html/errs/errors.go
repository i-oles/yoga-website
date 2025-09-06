package viewErrs

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	basicMessageErr = "Coś poszło nie tak, skontaktuj się ze mną."
)

func ErrBadRequest(c *gin.Context, tmplName string, err error) {
	slog.Error("handlerBadRequestErr", slog.String("error", err.Error()))
	c.HTML(http.StatusBadRequest, tmplName, gin.H{
		"Error": basicMessageErr,
	})
}

func ErrDTOConversion(c *gin.Context, tmplName string, err error) {
	slog.Error("handlerDTOConversionErr", slog.String("error", err.Error()))
	c.HTML(http.StatusInternalServerError, tmplName, gin.H{
		"Error": basicMessageErr,
	})
}

type IErrorHandler interface {
	Handle(ctx *gin.Context, tmplName string, err error)
}
