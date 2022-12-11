package main

import (
	"os"
	"regexp"

	"go.uber.org/zap"
)

func main() {
	initLog()

	toDel := []string{
		"swagger/cmd",
		"swagger/models",
		"swagger/restapi/operations",
		"swagger/restapi/doc.go",
		"swagger/restapi/embedded_spec.go",
		"swagger/restapi/server.go",
		"models",
		"mes",
	}

	if err := delete(toDel); err != nil {
		zap.L().Warn("fail to delete files", zap.Error(err))
	}
}

func initLog() {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
}

func delete(toDel []string) error {
	isFile := regexp.MustCompile(`(?m).*(\.\w*)$`)
	for _, d := range toDel {
		if isFile.MatchString(d) {
			if err := deleteFile(d); err != nil {
				return err
			}
			continue
		}

		if err := deleteDirectory(d); err != nil {
			return err
		}
	}
	return nil
}

func deleteFile(f string) error {
	zap.L().Info("delete file", zap.String("file", f))
	return os.Remove(f)
}

func deleteDirectory(d string) error {
	zap.L().Info("delete directory", zap.String("dir", d))
	return os.RemoveAll(d)
}
