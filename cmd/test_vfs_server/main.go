package main

import (
	"context"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"net/http"
	"time"
)

var (
	testEndpoint  = utils.GetEnv("MINIO_ENDPOINT", "localhost:9000")
	testAccessKey = utils.GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	testSecretKey = utils.GetEnv("MINIO_SECRET_KEY", "minioadmin")
	testUseSSL    = utils.GetEnvBool("MINIO_USE_SSL", false)
)

func main() {
	ctx := context.Background()

	cfg := &lib.VfsConfig{
		Endpoint:    testEndpoint,
		AccessKeyID: testAccessKey,
		SecretKey:   testSecretKey,
		Region:      "",
		Bucket:      "html-inline-test-" + time.Now().Format("20060102150405"),
		UseSSL:      testUseSSL,
		CACert:      "",
	}
	htmlContent := []byte(`<html><body><h1>Hello, Proxy!</h1></body></html>`)
	htmlFilenames := []string{"page.html", "page.htm", "page.png"}
	fmt.Printf("using bucket: %s\n", cfg.Bucket)

	vfs, err := lib.NewVfs(cfg)
	if err != nil {
		panic(err)
	}
	for _, filename := range htmlFilenames {
		objectPath := "html/" + filename

		err = vfs.Write(ctx, objectPath, htmlContent)
	}
	proxyHandler, err := vfs.Proxy("/public/", "/")

	if err := http.ListenAndServe(":7070", proxyHandler); err != nil {
		panic(err)
	}
}
