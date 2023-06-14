package ding

import (
	dingding2 "ding/model/dingding"
	"github.com/gin-gonic/gin"
)

func SubscribeToSomeone(c *gin.Context) {
	relationship := &dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	relationship.SubscribeSomeone()
}
func Unsubscribe(c *gin.Context) {
	relationship := dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	relationship.UnsubscribeSomeone()
}
