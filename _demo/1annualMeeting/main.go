/**
 * curl http://localhost:8080/
 * curl --data "users=weijie,chenxi,yangbo,yangkainan" http://localhost:8080/import
 * curl http://localhost:8080/lucky
 */
package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"math/rand"
	"strings"
	"time"
)

var userList []string

type lotteryController struct {
	Ctx iris.Context
}

func newApp() *iris.Application {
	application := iris.New()
	mvc.New(application.Party("/")).Handle(&lotteryController{})
	return application
}

func main() {
	app := newApp()
	userList = []string{}

	err := app.Run(iris.Addr(":8080"))
	if err != nil {
		return
	}
}

func (c *lotteryController) Get() string {
	count := len(userList)
	return fmt.Sprintf("当前总共参与抽奖的人数：%d\n", count)
}

// PostImport POST http://localhost:8080/import
// params: users
func (c *lotteryController) PostImport() string {
	strUsers := c.Ctx.FormValue("users")  // "users=zhaokai,liyuan"
	users := strings.Split(strUsers, ",") // users为 [zhaokai liyuan]
	count1 := len(userList)
	for _, u := range users {
		u = strings.TrimSpace(u)
		if len(u) > 0 {
			userList = append(userList, u)
		}
	}
	count2 := len(userList)
	return fmt.Sprintf("当前总共参与抽奖的用户数：%d, 成功导入的用户数：%d\n", count2, count2-count1)
}

// GetLucky GET http://localhost:8080/lucky
func (c *lotteryController) GetLucky() string {
	count := len(userList)
	if count > 1 {
		seed := time.Now().UnixNano()
		//Int31n 用于返回一个类型为 int32 的伪随机非负整数, 其值属于左闭右开区间 [0, n), 其中 n 即调用该函数时传入的参数
		index := rand.New(rand.NewSource(seed)).Int31n(int32(count))
		user := userList[index]
		//append，相当于做了一个拼接把第二个参数追加到第一个参数后面，但是如果第二个参数是一个切片，需要加三个点
		userList = append(userList[0:index], userList[index+1:]...)
		return fmt.Sprintf("当前中奖用户：%s, 剩余用户数：%d\n", user, count-1)
	} else if count == 1 {
		user := userList[0]
		return fmt.Sprintf("当前中奖用户：%s, 剩余用户数：%d\n", user, 0)
	} else {
		return fmt.Sprintf("已经没有参与用户，请先通过 /import 导入用户\n")
	}
}
