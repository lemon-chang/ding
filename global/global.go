package global

import (
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var (
	GLOBAL_GIN_Engine *gin.Engine
	GLOAB_DB          *gorm.DB      //mysql数据库连接
	GLOBAL_REDIS      *redis.Client //redis连接
	GLOAB_CORN        *cron.Cron    //Cron定时器连接
	GLOAB_VALIDATOR   *validator.Validate
	GLOBAL_Feishu     *lark.Client        //飞书客户端
	GLOBAL_Kafka_Prod sarama.SyncProducer //kafka生产者
	GLOBAL_Kafka_Cons sarama.Consumer     //kafka消费者
)

// KafMsg 封装kaf消息
func KafMsg(topic, con string, partition int32) *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(con),
		Partition: partition,
	}
}
