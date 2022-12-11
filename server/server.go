package server

import (
	"context"
	"errors"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomImpl "gitlab.kenda.com.tw/kenda/mcom/impl"

	"gitlab.kenda.com.tw/kenda/mui/server/configs"
)

var (
	// ErrorNilConnection for nil data manager error message definition.
	ErrorNilConnection = errors.New("nil data manager connection")
)

// RegisterDataManager registers data manager.
func RegisterDataManager(configs configs.Configs) (mcom.DataManager, error) {
	options := []mcomImpl.Option{
		mcomImpl.WithPostgreSQLSchema(configs.PostgreSQL.Schema),
		mcomImpl.WithPDAWebServiceEndpoint(configs.WebServiceEndpoint),
	}

	if !configs.ActiveDirectory.IsEmpty() {
		options = append(options, mcomImpl.ADAuth(mcomImpl.ADConfig{
			Host:          configs.ActiveDirectory.Host,
			Port:          configs.ActiveDirectory.Port,
			DN:            configs.ActiveDirectory.DN,
			QueryUser:     configs.ActiveDirectory.QueryUser,
			QueryPassword: configs.ActiveDirectory.QueryPassword,
			WithTLS:       configs.ActiveDirectory.WithTLS,
		}))
	}

	return mcomImpl.New(
		context.Background(),
		mcomImpl.PGConfig{
			Database: configs.PostgreSQL.Name,
			Address:  configs.PostgreSQL.Address,
			Port:     configs.PostgreSQL.Port,
			UserName: configs.PostgreSQL.UserName,
			Password: configs.PostgreSQL.Password,
		},
		options...,
	)
}
