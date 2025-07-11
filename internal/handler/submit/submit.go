package submit

import (
	"errors"
	"main/internal/generator"
	"main/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ClassesRepo       repository.Classes
	PractitionersRepo repository.Practitioners
	TokenGenerator    generator.Token
}

func NewHandler(
	classesRepo repository.Classes,
	practitionersRepo repository.Practitioners,
	tokenGenerator generator.Token,
) *Handler {
	return &Handler{
		ClassesRepo:       classesRepo,
		PractitionersRepo: practitionersRepo,
		TokenGenerator:    tokenGenerator,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	classIDStr := c.PostForm("classID")

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	class, err := h.ClassesRepo.Get(classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

		return
	}

	if class.SpotsLeft == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("this class is fully booked")})
	}

	name := c.PostForm("name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	err = h.PractitionersRepo.Insert(ctx, classID, name, lastName, email)
	if err != nil {
		if strings.Contains(err.Error(), "already booked") {
			c.HTML(http.StatusConflict, "book.tmpl", gin.H{
				"ID":    classID,
				"Error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	//err = h.ClassesRepo.Update(classID)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//
	//	return
	//}

	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}
