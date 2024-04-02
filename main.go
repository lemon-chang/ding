package main

import (
	"context"
	"ding/initialize/enter"
	"ding/initialize/viper"
	"ding/routers"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	enter.Init()
	r := routers.Setup(viper.Conf.Mode)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.Conf.App.Port),
		Handler: r,
	}
	// 初始化kafka
	//if err = initialize.KafkaInit(); err != nil {
	//	zap.L().Error(fmt.Sprintf("kafka init failed ... ,err:%v\n", err))
	//}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("lister: %s\n", err)
			return
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("Server Shutdown", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
