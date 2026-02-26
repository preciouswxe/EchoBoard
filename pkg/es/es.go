package es

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/preciouswxe/EchoBoard_backend/setting"
	"go.uber.org/zap"
)

var Client *elasticsearch.TypedClient

func Init(cfg *setting.EsConfig) (err error) {
	Client, err = elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		},
	})
	if err != nil {
		zap.L().Error("elasticsearch.NewTypedClient failed, err:", zap.Error(err))
		return err
	}
	zap.L().Info("ES init success")

	return nil
}

