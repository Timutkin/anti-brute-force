package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/timutkin/anti-brute-force/internal/service"
	"net/http"
)

type Response struct {
	Err string
}

type ListService interface {
	AddCIDR(ipAndMask string) error
	DeleteCIDR(ipAndMask string) error
	GetCIDRs() []string
}

type ListHandler struct {
	service ListService
}

func NewListHandler(service ListService) ListHandler {
	return ListHandler{service: service}
}

func (b ListHandler) AddCIDR() func(c *gin.Context) {
	return func(c *gin.Context) {
		b.command(c, func(cidr string) error {
			return b.service.AddCIDR(cidr)
		})
	}
}

func (b ListHandler) DeleteCIDR() func(c *gin.Context) {
	return func(c *gin.Context) {
		b.command(c, func(cidr string) error {
			return b.service.DeleteCIDR(cidr)
		})
	}
}

func (b ListHandler) GetCIDRs() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, b.service.GetCIDRs())
	}
}

func (b ListHandler) command(c *gin.Context, commandFunc func(cidr string) error) {
	cidr := c.Query("cidr")
	if cidr == "" {
		c.JSON(http.StatusBadRequest, Response{Err: "not found cidr query param"})
		return
	}
	err := commandFunc(cidr)
	if err != nil {
		if errors.Is(err, service.ErrParseCIDR) {
			c.JSON(http.StatusBadRequest, Response{Err: err.Error()})
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
