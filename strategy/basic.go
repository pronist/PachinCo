package strategy

import (
	"fmt"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

// 단순한 분할 매수 전략이다. 싸게사서 조금 더 고가에 파는 기본에 충실한 전략이다.
// 매도보다 매수를 더 우선으로 처리한다.
//
// 코인이 없을 때
// * 실시간 가격이 전일 `종가`보다 n% 하락하면 매수
// * 실시간 가격이 직전 `매도`보다 n% 하락하면 매수
// 코인이 있을 때
// * 실시간 가격이 매수 `평균`보다 n% 하락하면 매수
// *** 매도
// * 실시간 가격이 매수 `평균`보다 n% 상승하면 매도
//
// *** 장점
// * 하락에 대해 일정 비율로 분산 매수하므로 평단가 낮추기에 기여한다.
// *** 단점
// * 매도에 대해서는 손절매를 하지 않고 기다리기 때문에 봇이 활동적으로 움직이지 않는다.
// * 매수에 상승을 포함한 급등주를 따라가지 않으므로 수익성은 낮다.
type Basic struct{}

func (b *Basic) Tracking(coins map[string]float64, coin string) {
	if r, ok := coins[coin]; ok {
		accounts, balances, limitOrderPrice, orderBuyingPrice := update(coin, r)

		for {
			price, err := upbit.API.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
			}

			if balances["KRW"] >= upbit.MinimumOrderPrice && balances["KRW"] > orderBuyingPrice && orderBuyingPrice > upbit.MinimumOrderPrice {
				volume := orderBuyingPrice / price

				if math.IsInf(volume, 0) {
					bot.LogChan <- bot.Log{Msg: "division by zero", Level: logrus.FatalLevel}
				}

				if coinBalance, ok := balances[coin]; ok {
					avgBuyPrice, err := upbit.API.GetAverageBuyPrice(accounts, coin)
					if err != nil {
						bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
					}

					p := price / avgBuyPrice

					////// 매수 전략

					// 1. 분할 매수 전략 (하락시 평균단가를 낮추는 전략)
					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우

					if math.IsInf(p, 0) {
						bot.LogChan <- bot.Log{Msg: "division by zero", Level: logrus.FatalLevel}
					}

					if avgBuyPrice*coinBalance+orderBuyingPrice <= limitOrderPrice {
						if p-1 <= upbit.Config.L {
							accounts, balances, limitOrderPrice, orderBuyingPrice = order(coin, "bid", r, volume, price)
							continue
						}
					}

					////// 매도 전략

					// 전량 매도
					// 더 높은 기대수익률을 바라보기 어려워짐. 하락 리스크에 조금 더 방어적이지만
					// 너무 수비적이라 조금 더 공격적으로 해도 될 것 같음.

					// 매도에는 하락장에 대한 전략이 없음. 오히려 하락하는 경우 추가 매수.

					orderSellingPrice := coinBalance * price

					if math.IsInf(p, 0) {
						bot.LogChan <- bot.Log{Msg: "division by zero", Level: logrus.FatalLevel}
					}

					// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
					if p-1 >= upbit.Config.H && orderSellingPrice > upbit.MinimumOrderPrice {
						accounts, balances, limitOrderPrice, orderBuyingPrice = order(coin, "ask", r, coinBalance, price)
						continue
					}
				} else {
					////// 코인을 처음 살 떄의 매수 전략
					var isBuyable bool

					// 이 경우는 매도 가격을 기준으로 한다.
					orders, err := upbit.API.GetOrderList("KRW-"+coin, "done")
					if err != nil {
						bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
					}

					// 이 매도에는 시장가 매도가 제외된다. 즉, 웹에서 시장가에 매도한 것이 아니라
					// 봇에서 지정가에 매도한 것만 처리된다.
					askOrders := upbit.API.GetAskOrders(orders)

					if len(askOrders) > 0 {
						latestAskPrice, err := upbit.API.GetLatestAskPrice(orders)
						if err != nil {
							bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
						}

						pp := price / latestAskPrice // 마지막 매도가 대비 변화율

						if math.IsInf(pp, 0) {
							bot.LogChan <- bot.Log{Msg: "division by zero", Level: logrus.FatalLevel}
						}

						// 마지막으로 매도한 가격을 기준으로 매수
						isBuyable = pp-1 <= upbit.Config.F
					}

					daysCandles, err := upbit.API.GetCandlesDays("KRW-"+coin, "1")
					if err != nil {
						bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
					}

					// 전날 또는 매도 이후 변동을 기준으로 매수
					if (daysCandles[0]["change_rate"].(float64) <= upbit.Config.F) || isBuyable {
						accounts, balances, limitOrderPrice, orderBuyingPrice = order(coin, "bid", r, volume, price)
						continue
					}
				}
			}

			time.Sleep(1 * time.Second)
		}
	} else {
		bot.LogChan <- bot.Log{
			Msg: fmt.Errorf("Not found coin '%s' in supported markets", coin), Level: logrus.FatalLevel,
		}
	}
}
