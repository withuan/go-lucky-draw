package main

import (
	"fmt"
	"github.com/kataras/iris/v12/httptest"
	"sync"
	"testing"
)

func TestMVC(t *testing.T) {
	e := httptest.New(t, newApp())

	var wg sync.WaitGroup //fixme
	e.GET("/").Expect().Status(httptest.StatusOK).Body().Equal("当前总共参与抽奖的人数：0\n")

	for i := 0; i < 100; i++ {
		wg.Add(1) //fixme
		go func(i int) {
			defer wg.Done()

			e.POST("/import").WithFormField("users", fmt.Sprintf("test_u%d", i)).Expect().
				Status(httptest.StatusOK)
		}(i)
	}

	wg.Wait()

	e.GET("/").Expect().Status(httptest.StatusOK).Body().Equal("当前总共参与抽奖的人数：100\n")
	e.GET("/lucky").Expect().Status(httptest.StatusOK)
	e.GET("/").Expect().Status(httptest.StatusOK).Body().Equal("当前总共参与抽奖的人数：99\n")
}
