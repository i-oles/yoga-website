package allbookingsforclass

import (
	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	bookingsRepo    repositories.IBookings
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	bookingsRepo repositories.IBookings,
	apiErrorHandler apiErrs.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsRepo:    bookingsRepo,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	classIDStr := c.Param("class_id")

	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	allBookingsForClass, err := h.bookingsRepo.GetAllByClassID(ctx, classID)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	response, err := dto.ToBookingsListResponse(allBookingsForClass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusOK, response)
}
