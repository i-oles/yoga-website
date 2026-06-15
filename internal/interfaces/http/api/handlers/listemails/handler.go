package listemails

import (
	"net/http"

	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type handler struct {
	bookingsRepo    repositories.IBookings
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	bookingsRepo repositories.IBookings,
	apiErrorHandler apiErrs.IErrorHandler,
) *handler {
	return &handler{
		bookingsRepo:    bookingsRepo,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *handler) Handle(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	allBookings, err := h.bookingsRepo.List(ctx)
	if err != nil {
		h.apiErrorHandler.Handle(ginCtx, err)

		return
	}

	uniqueEmails := make(map[string]struct{}, len(allBookings))

	for _, booking := range allBookings {
		uniqueEmails[booking.Email] = struct{}{}
	}

	emails := make([]string, 0, len(uniqueEmails))

	for email := range uniqueEmails {
		emails = append(emails, email)
	}

	ginCtx.JSON(http.StatusOK, dto.BookingEmailsResponse{
		Emails: emails,
	})
}
