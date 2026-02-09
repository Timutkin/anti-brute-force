package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timutkin/anti-brute-force/internal/service"
)

type BruteForceService interface {
	IsCredentialAllowed(login, password, ip string) (bool, error)
	GetLoginBuckets() []service.BucketsView
	DeleteBuckets(login, ip string)
}

type BruteForceHandler struct {
	bruteForce BruteForceService
}

func NewBruteForceHandler(bruteForce BruteForceService) BruteForceHandler {
	return BruteForceHandler{bruteForce}
}

type AllowAuthorizationRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"ip" binding:"required"`
}

func (b *BruteForceHandler) AllowAuthorization() func(c *gin.Context) {
	return func(c *gin.Context) {
		var rq AllowAuthorizationRequest
		err := c.ShouldBind(&rq)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{Err: err.Error()})
			return
		}
		result, err := b.bruteForce.IsCredentialAllowed(rq.Login, rq.Password, rq.IP)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if result {
			c.Writer.WriteHeader(http.StatusOK)
			return
		}
		c.Writer.WriteHeader(http.StatusTooManyRequests)
	}
}

func (b *BruteForceHandler) GetBuckets() func(c *gin.Context) {
	return func(c *gin.Context) {
		buckets := b.bruteForce.GetLoginBuckets()
		c.JSON(http.StatusOK, buckets)
	}
}

func (b *BruteForceHandler) DeleteBuckets() func(c *gin.Context) {
	return func(c *gin.Context) {
		loginParam := c.Query("login")
		ipParam := c.Query("ip")
		if loginParam == "" && ipParam == "" {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		b.bruteForce.DeleteBuckets(loginParam, ipParam)
		c.Writer.WriteHeader(http.StatusOK)
	}
}
