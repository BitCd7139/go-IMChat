package ssl

import (
	"IMChat/pkg/zlog"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"strconv"
)

func TlsHandler(host string, prot int) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     host + ":" + strconv.Itoa(prot),
		})

		err := secureMiddleware.Process(c.Writer, c.Request)

		if err != nil {
			zlog.Fatal(err.Error())
			return
		}

		c.Next()
	}

}
