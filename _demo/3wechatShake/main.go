/**
 *	微信摇一摇
 *	基础功能：/lucky 只有一个抽奖的接口
 */
package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"log"
	"math/rand"
	"os"
	"time"
)

// 奖品类型
const (
	giftTypeCoin      = iota //虚拟币
	giftTypeCoupon           //不同券
	giftTypeCouponFix        //相同的券
	giftTypeRealSmall        //实物小奖
	giftTypeRealLarge        //实物大奖
)

type gift struct {
	id       int      //奖品ID
	name     string   //奖品名称
	pic      string   //奖品图片
	link     string   //奖品链接
	gtype    int      //奖品类型
	data     string   //奖品的数据（特定的配置信息）
	datalist []string //奖品数据集合（不同优惠券的编码）
	total    int      //总数， 0 不限量
	left     int      //剩余数量
	inuse    bool     //是否使用中
	rate     int      //中奖概率，万分之N，0-9999
	rateMin  int      //大于等于最小中奖编码
	rateMax  int      //小于中奖编码
}

//最大的中奖号码
const rateMax = 10000

var logger *log.Logger

//奖品列表
var giftList []*gift

type lotteryController struct {
	Ctx iris.Context
}

//初始化日志
func initLog() {
	f, _ := os.Create("/var/log/lottery_demo.log")
	logger = log.New(f, "", log.Ldate|log.Lmicroseconds)
}

//初始化奖品列表
func initGift() {
	giftList = make([]*gift, 5)
	g1 := gift{
		id:       1,
		name:     "手机大奖",
		pic:      "",
		link:     "",
		gtype:    giftTypeRealLarge,
		data:     "",
		datalist: nil,
		total:    10,
		left:     10,
		inuse:    true,
		rate:     10,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[0] = &g1

	g2 := gift{
		id:       2,
		name:     "充电器",
		pic:      "",
		link:     "",
		gtype:    giftTypeRealSmall,
		data:     "",
		datalist: nil,
		total:    5,
		left:     5,
		inuse:    true,
		rate:     100,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[1] = &g2

	g3 := gift{
		id:       3,
		name:     "优惠券满200减50元",
		pic:      "",
		link:     "",
		gtype:    giftTypeCouponFix,
		data:     "mall-coupon-2018",
		datalist: nil,
		total:    5,
		left:     5,
		inuse:    true,
		rate:     5000,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[2] = &g3

	g4 := gift{
		id:       4,
		name:     "直降优惠券50元",
		pic:      "",
		link:     "",
		gtype:    giftTypeCoupon,
		data:     "",
		datalist: []string{"c01", "c02", "c03", "c04", "c05"},
		total:    5,
		left:     5,
		inuse:    true,
		rate:     2000,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[3] = &g4

	g5 := gift{
		id:       5,
		name:     "金币",
		pic:      "",
		link:     "",
		gtype:    giftTypeCoin,
		data:     "10金币",
		datalist: nil,
		total:    5,
		left:     5,
		inuse:    true,
		rate:     5000,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[4] = &g5

	//数据整理，中奖区间数据
	rateStart := 0
	for _, data := range giftList {
		if !data.inuse {
			continue
		}
		data.rateMin = rateStart
		data.rateMax = rateStart + data.rate
		if data.rateMax >= rateMax {
			data.rateMax = rateMax
			rateStart = 0
		} else {
			rateStart += data.rate
		}
	}
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})

	initLog()
	initGift()
	return app
}

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

// Get 奖品数量的信息 GET http://localhost:8080/
func (c *lotteryController) Get() string {
	count := 0
	total := 0
	for _, data := range giftList {
		if data.inuse && (data.total == 0 ||
			(data.total > 0 && data.left > 0)) {
			count++
			total += data.left
		}
	}
	return fmt.Sprintf("当前有效奖品种类数量：%d,限量奖品数量=%d\n", count, total)
}

// GetLucky 抽奖 GET http://localhost:8080/lucky
func (c *lotteryController) GetLucky() map[string]interface{} {
	code := luckyCode()
	ok := false
	result := make(map[string]interface{})
	result["success"] = false

	for _, data := range giftList {
		if !data.inuse || (data.total > 0 && data.left <= 0) {
			continue
		}
		if data.rateMin <= int(code) && data.rateMax > int(code) {
			// 中奖了，抽奖编码在奖品编码范围内
			// 开始发奖
			sendData := ""
			switch data.gtype {
			case giftTypeCoin:
				ok, sendData = sendCoin(data)
			case giftTypeCoupon:
				ok, sendData = sendCoupon(data)
			case giftTypeCouponFix:
				ok, sendData = sendCouponFix(data)
			case giftTypeRealSmall:
				ok, sendData = sendRealSmall(data)
			case giftTypeRealLarge:
				ok, sendData = sendRealLarge(data)
			}
			if ok {
				//中奖后，成功得到奖品
				//生成中奖记录
				saveLuckyData(code, data.id, data.name, data.link, sendData, data.left)
				result["success"] = ok
				result["id"] = data.id
				result["name"] = data.name
				result["link"] = data.link
				result["data"] = sendData
				break
			}
		}
	}

	return result
}

func luckyCode() int32 {
	seed := time.Now().UnixNano()
	code := rand.New(rand.NewSource(seed)).Int31n(int32(rateMax)) // Int31n 返回 [0,n)
	return code                                                   // 0 - 9999
}

func sendCoin(data *gift) (bool, string) {
	if data.total == 0 {
		//数量无限
		return true, data.data
	} else if data.left > 0 {
		//还有剩余
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

func sendCoupon(data *gift) (bool, string) {
	if data.total == 0 {
		//数量无限
		return true, data.data
	} else if data.left > 0 {
		//还有剩余
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

func sendCouponFix(data *gift) (bool, string) {
	if data.total == 0 {
		//数量无限
		return true, data.data
	} else if data.left > 0 {
		//还有剩余
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

func sendRealSmall(data *gift) (bool, string) {
	if data.total == 0 {
		//数量无限
		return true, data.data
	} else if data.left > 0 {
		//还有剩余
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

func sendRealLarge(data *gift) (bool, string) {
	if data.total == 0 {
		//数量无限
		return true, data.data
	} else if data.left > 0 {
		//还有剩余
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

//记录用户的获奖信息
func saveLuckyData(code int32, id int, name, link, sendData string, left int) {
	logger.Printf("Lucky, code=%d, gift=%d, name=%s, link=%s, data=%s, left=%d \n",
		code, id, name, link, sendData, left)
}
