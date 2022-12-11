package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Options for the implementation
type Options struct {
	ServerConfig string `short:"c" long:"server-config" description:"server configuration file" required:"true"`
}

// DBConnection set PostgreSQL connection settings.
type DBConnection struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
}

// ActiveDirectory definition.
type ActiveDirectory struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	DN            string `yaml:"base_dn"`
	QueryUser     string `yaml:"query_user"`
	QueryPassword string `yaml:"query_password"`
	WithTLS       bool   `yaml:"with_tls"`
}

// UnmarshalYAML unmarshal the yaml document and replace ${var} or $var
// in the string according to the values of the current environment variables.
func (ad *ActiveDirectory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tempActiveDirectory struct {
		Host          string      `yaml:"host"`
		Port          interface{} `yaml:"port"`
		DN            string      `yaml:"base_dn"`
		QueryUser     string      `yaml:"query_user"`
		QueryPassword string      `yaml:"query_password"`
		WithTLS       bool        `yaml:"with_tls"`
	}
	var t tempActiveDirectory
	if err := unmarshal(&t); err != nil {
		return err
	}

	port := 0
	switch type_ := t.Port.(type) {
	case int8:
		port = int(type_)
	case int16:
		port = int(type_)
	case int32:
		port = int(type_)
	case int64:
		port = int(type_)
	case int:
		port = type_
	case string:
		type_ = os.ExpandEnv(type_)

		i, err := strconv.Atoi(type_)
		if err != nil {
			return fmt.Errorf("invalid active directory port: %v", err)
		}
		port = i
	default:
		return fmt.Errorf("invalid type: %T", type_)
	}

	*ad = ActiveDirectory{
		Host:          os.ExpandEnv(t.Host),
		Port:          port,
		DN:            os.ExpandEnv(t.DN),
		QueryUser:     os.ExpandEnv(t.QueryUser),
		QueryPassword: os.ExpandEnv(t.QueryPassword),
		WithTLS:       t.WithTLS,
	}

	return nil
}

func (ad *DBConnection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tempDBConnection struct {
		Name     string      `yaml:"name"`
		Address  string      `yaml:"address"`
		Port     interface{} `yaml:"port"`
		UserName string      `yaml:"username"`
		Password string      `yaml:"password"`
		Schema   string      `yaml:"schema"`
	}
	var t tempDBConnection
	if err := unmarshal(&t); err != nil {
		return err
	}

	port := 0
	switch type_ := t.Port.(type) {
	case int8:
		port = int(type_)
	case int16:
		port = int(type_)
	case int32:
		port = int(type_)
	case int64:
		port = int(type_)
	case int:
		port = type_
	case string:
		type_ = os.ExpandEnv(type_)

		i, err := strconv.Atoi(type_)
		if err != nil {
			return fmt.Errorf("invalid PostgreSQL port: %v", err)
		}
		port = i
	default:
		return fmt.Errorf("invalid type: %T", type_)
	}

	*ad = DBConnection{
		Name:     os.ExpandEnv(t.Name),
		Address:  os.ExpandEnv(t.Address),
		Port:     port,
		UserName: os.ExpandEnv(t.UserName),
		Password: os.ExpandEnv(t.Password),
		Schema:   os.ExpandEnv(t.Schema),
	}
	return nil
}

// IsEmpty checks if ad configuration is set
func (ad ActiveDirectory) IsEmpty() bool {
	return ad.Host == "" &&
		ad.Port == 0 &&
		ad.DN == "" &&
		ad.QueryUser == "" &&
		ad.QueryPassword == ""
}

type FunctionAPIPath struct {
	LoadWorkOrderAPIPath   string `yaml:"loadWorkOrder"`
	ClosedWorkOrderAPIPath string `yaml:"closedWorkOrder"`
	BindResourceAPIPath    string `yaml:"bindResource"`
}

// Configs for
type Configs struct {
	DevMode        bool   `yaml:"development_mode"`
	UIDir          string `yaml:"ui_distribution_directory"`
	CreateUIConfig bool   `yaml:"create_ui_configuration"`

	Timeout                 time.Duration              `yaml:"timeout"`
	WebServiceEndpoint      string                     `yaml:"web_service_endpoint"`
	PostgreSQL              DBConnection               `yaml:"postgres"`
	ActiveDirectory         ActiveDirectory            `yaml:"active_directory"`
	CorsAllowedOrigins      []string                   `yaml:"cors_allowed_origins"`
	TokenExpiredSeconds     int                        `yaml:"token_expired_in_seconds"`
	FunctionRolePermissions map[string][]string        `yaml:"permissions"`
	Printers                map[string]string          `yaml:"printers"`
	FontPath                string                     `yaml:"font_path"`
	StationFunctionConfig   map[string]FunctionAPIPath `yaml:"station_function_config"`
	MesPath                 string                     `yaml:"mes_path"`
}
