package ding

import (
	dingding2 "ding/model/dingding"
	"ding/response"
	"github.com/gin-gonic/gin"
)

func SubscribeToSomeone(c *gin.Context) {
	relationship := &dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	err := relationship.SubscribeSomeone()
	if err != nil {
		response.FailWithMessage("", c)
	} else {
		response.OkWithMessage("", c)
	}
}

func Unsubscribe(c *gin.Context) {
	relationship := dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	err := relationship.UnsubscribeSomeone()
	if err != nil {
		response.FailWithMessage("", c)
	} else {
		response.OkWithMessage("", c)
	}
}
