package configs

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestActiveDirectory_UnmarshalYAML(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(os.Getenv, func(key string) string {
		switch key {
		case "TEST_HOST":
			return "my-host"
		case "TEST_PORT":
			return "1234"
		case "TEST_BASE_DN":
			return "my-base-dn"
		case "TEST_QUERY_USER":
			return "my-query-user"
		case "TEST_PASSWORD":
			return "my-password"
		}
		return ""
	})
	defer monkey.UnpatchAll()

	{ // integer type port
		text := `
host: my-host
port: 1234
base_dn: my-base-dn
query_user: my-query-user
query_password: my-password
with_tls: true`

		var ad ActiveDirectory
		assert.NoError(yaml.Unmarshal([]byte(text), &ad))
		assert.Equal(ActiveDirectory{
			Host:          "my-host",
			Port:          1234,
			DN:            "my-base-dn",
			QueryUser:     "my-query-user",
			QueryPassword: "my-password",
			WithTLS:       true,
		}, ad)
	}
	{ // bad case: string type port
		text := `
host: my-host
port: TEST_PORT
base_dn: my-base-dn
query_user: my-query-user
query_password: my-password
with_tls: true`

		var ad ActiveDirectory
		assert.Error(yaml.Unmarshal([]byte(text), &ad))
	}
	{ // values from environment variable
		text := `
host: ${TEST_HOST}
port: ${TEST_PORT}
base_dn: ${TEST_BASE_DN}
query_user: ${TEST_QUERY_USER}
query_password: ${TEST_PASSWORD}
with_tls: true`

		var ad ActiveDirectory
		assert.NoError(yaml.Unmarshal([]byte(text), &ad))
		assert.Equal(ActiveDirectory{
			Host:          "my-host",
			Port:          1234,
			DN:            "my-base-dn",
			QueryUser:     "my-query-user",
			QueryPassword: "my-password",
			WithTLS:       true,
		}, ad)
	}
}
func TestDBConnection_UnmarshalYAML(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(os.Getenv, func(key string) string {
		switch key {
		case "TEST_NAME":
			return "my-name"
		case "TEST_ADDRESS":
			return "my-address"
		case "TEST_PORT":
			return "1234"
		case "TEST_USERNAME":
			return "my-username"
		case "TEST_PASSWORD":
			return "my-password"
		case "TEST_SCHEMA":
			return "my-schema"
		}
		return ""
	})
	defer monkey.UnpatchAll()

	{ // integer type port
		text := `
name: my-name
address: my-address
port: 1234
username: my-username
password: my-password
schema: my-schema`

		var db DBConnection
		assert.NoError(yaml.Unmarshal([]byte(text), &db))
		assert.Equal(DBConnection{
			Name:     "my-name",
			Address:  "my-address",
			Port:     1234,
			UserName: "my-username",
			Password: "my-password",
			Schema:   "my-schema",
		}, db)
	}
	{ // bad case: string type port
		text := `
name: my-name
address: my-address
port: TEST_PORT
username: my-username
password: my-password
schema: my-schema`

		var db DBConnection
		assert.Error(yaml.Unmarshal([]byte(text), &db))
	}
	{ // values from environment variable
		text := `
name: ${TEST_NAME}
address: ${TEST_ADDRESS}
port: ${TEST_PORT}
username: ${TEST_USERNAME}
password: ${TEST_PASSWORD}
schema: ${TEST_SCHEMA}`

		var db DBConnection
		assert.NoError(yaml.Unmarshal([]byte(text), &db))
		assert.Equal(DBConnection{
			Name:     "my-name",
			Address:  "my-address",
			Port:     1234,
			UserName: "my-username",
			Password: "my-password",
			Schema:   "my-schema",
		}, db)
	}
}
