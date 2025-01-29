package koko

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/exchange"
	"github.com/jumpserver/koko/pkg/httpd"
	"github.com/jumpserver/koko/pkg/i18n"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"

	"github.com/jumpserver/koko/pkg/jms-sdk-go/model"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"
)

type Koko struct {
	webSrv *httpd.Server
	sshSrv *sshd.Server
}

func (k *Koko) Start() {
	go k.webSrv.Start()
	go k.sshSrv.Start()
}

func (k *Koko) Stop() {
	k.sshSrv.Stop()
	k.webSrv.Stop()
	logger.Info("Quit The KoKo")
}

func RunForever(confPath string) {
	config.Setup(confPath)
	bootstrap()
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	jmsService := MustJMService()
	//
	bootstrapWithJMService(jmsService)
	//
	webSrv := httpd.NewServer(jmsService)
	sshSrv := sshd.NewSSHServer(jmsService)
	app := &Koko{
		webSrv: webSrv,
		sshSrv: sshSrv,
	}
	app.Start()
	// runTasks(jmsService) 附加任务先去掉
	<-gracefulStop
	app.Stop()
}

func bootstrap() {
	i18n.Initial()
	logger.Initial()
}

func bootstrapWithJMService(jmsService *service.JMService) {
	updateEncryptConfigValue(jmsService)
	// redis
	exchange.Initial()
}

func updateEncryptConfigValue(jmsService *service.JMService) {
	cfg := config.GlobalConfig
	encryptKey := cfg.SecretEncryptKey
	if encryptKey != "" {
		redisPassword := cfg.RedisPassword
		ret, err := jmsService.GetEncryptedConfigValue(encryptKey, redisPassword)
		if err != nil {
			logger.Error("Get encrypted config value failed: " + err.Error())
			return
		}
		if ret.Value != "" {
			logger.Info(">>> updateEncryptConfigValue: " + ret.Value)
			cfg.UpdateRedisPassword(ret.Value)
		} else {
			logger.Error("Get encrypted config value failed: empty value")
		}
	}
}

func runTasks(jmsService *service.JMService) {
	logger.Info(">>> runTasks")
	if config.GetConf().UploadFailedReplay {
		go uploadRemainReplay(jmsService)
	}
	if config.GetConf().UploadFailedFTPFile {
		go uploadRemainFTPFile(jmsService)
	}
	// 和主服务的心跳
	go keepHeartbeat(jmsService)

	go RunConnectTokensCheck(jmsService)
}

func MustJMService() *service.JMService {
	key := MustLoadValidAccessKey()
	jmsService, err := service.NewAuthJMService(
		service.JMSCoreHost(config.GlobalConfig.CoreHost),
		service.JMSTimeOut(30*time.Second),
		service.JMSAccessKey(key.ID, key.Secret),
	)
	if err != nil {
		logger.Fatal("创建JMS Service 失败 " + err.Error())
		os.Exit(1)
	}
	return jmsService
}

func MustLoadValidAccessKey() model.AccessKey {
	var key = model.AccessKey{
		ID:     "b24cefff-dbd0-4f55-a85b-3cc3c59e7f05",
		Secret: "0nXhBpNdcIeJEU4o5pks2iG5XnAoLZ5lRNED",
	}
	// 校验accessKey
	return key
}
