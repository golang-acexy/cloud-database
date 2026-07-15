package mongo

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := loader.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动 Mongo Starter 失败: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	_, err := loader.StopAllBySetting(10 * time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "停止 Mongo Starter 失败: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
