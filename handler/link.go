package handler

import (
	"net/http"
	"url-shortener/service"

	"github.com/gin-gonic/gin"
)

type QueryJSON struct {
	URL string `json:"url"`
}

type LinkHandler struct {
	service *service.LinkService
}

func NewLinkHandler(s *service.LinkService) *LinkHandler {
	return &LinkHandler{service: s}
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	ctx := c.Request.Context()

	var linkQuery QueryJSON

	if err := c.ShouldBindJSON(&linkQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.service.CreateLink(ctx, linkQuery.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, link)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	ctx := c.Request.Context()

	short := c.Param("short")

	if err := h.service.DeleteLink(ctx, short); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) GetLink(c *gin.Context) {
	ctx := c.Request.Context()

	short := c.Param("short")

	link, err := h.service.GetLink(ctx, short)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) GetLinks(c *gin.Context) {
	ctx := c.Request.Context()

	links, err := h.service.GetLinks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, links)
}

func (h *LinkHandler) GetClicks(c *gin.Context) {
	ctx := c.Request.Context()

	short := c.Param("short")

	clicks, err := h.service.GetClicks(ctx, short)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clicks": clicks})
}
