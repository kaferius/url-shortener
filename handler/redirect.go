package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *LinkHandler) Redirect(c *gin.Context) {
	ctx := c.Request.Context()
	short := c.Param("short")
	link, err := h.service.GetLink(ctx, short)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusMovedPermanently, link.Long)
}
