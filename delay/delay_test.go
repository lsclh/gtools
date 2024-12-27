package delay

import (
	"fmt"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	RegisterMethod("hello", func(params string) {
		fmt.Println("开始时间" + params)
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "hello world")
	})

	t1 := New("hello",
		WithParams(time.Now().Format("2006-01-02 15:04:05")+" 10秒"),
		WithName("helloworldTask"),
	)
	t1.AddForAfter(time.Second * 5)

	t2 := New("hello",
		WithParams(time.Now().Format("2006-01-02 15:04:05")+" 15秒"),
	)
	t2.AddForTime(time.Now().Unix() + 7)
	time.Sleep(time.Second * 2)

	Del("helloworldTask")
	select {}
}
