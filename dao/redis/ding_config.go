package redis

import (
	"context"
	"ding/global"
	"ding/utils"
	"go.uber.org/zap"
)

func GetCropId() (CropId string, err error) {
	CropId, err = global.GLOBAL_REDIS.Get(context.Background(), utils.CropId).Result()
	if err != nil {
		zap.L().Error("从redis从取CropId失败", zap.Error(err))
		return
	}
	return
}


