package httprpc

import (
	"log"
	"time"
)

//AccessLog access request log
func AccessLog(next HandlerFunc) HandlerFunc {
	return func(c *Context) (err error) {
		url := c.Request().URL
		begin := time.Now()
		err = next(c)
		if err != nil {
			log.Printf("[error] request %s error,err:%s\n", url.String(), err.Error())
		} else {
			log.Printf("[info] request %s success,cost %v\n", url.String(), time.Since(begin))
		}
		return
	}
}
