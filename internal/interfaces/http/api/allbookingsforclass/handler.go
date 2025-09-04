package allbookingsforclass

import (
	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	"main/internal/interfaces/http/err/handler"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	bookingsRepo repositories.IBookings
	errorHandler handler.IErrorHandler
}

func NewHandler(
	bookingsRepo repositories.IBookings,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsRepo: bookingsRepo,
		errorHandler: errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var req dto.GetAllBookingsForClassRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx := c.Request.Context()

	allBookingsForClass, err := h.bookingsRepo.GetAllByClassID(ctx, classID)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	response, err := dto.ToBookingsListResponse(allBookingsForClass)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	c.JSON(http.StatusOK, response)
}
