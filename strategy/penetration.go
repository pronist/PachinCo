package strategy

// 변동성 돌파전략이다. 상승장에 구입한다.
type Penetration struct {
	H float64 // 판매 상승 기준
	K float64 // 돌파 상수
}

// 변동성 돌파 되었는지 검사
//func (p *Penetration) Is(dayCandles []map[string]interface{}, price float64) bool {
//	R := dayCandles[1]["high_price"].(float64) - dayCandles[1]["low_price"].(float64)
//	return dayCandles[0]["opening_price"].(float64) + (R * p.K) < price
//}
//
//func (p *Penetration) Run(coin *bot.Coin) {
//	bot.LogChan <- bot.Log{
//		Msg: "Start PENETRATION strategy ...",
//		Fields: logrus.Fields{
//			"coin": coin.Name,
//		},
//		Level: logrus.InfoLevel,
//	}
//
//	err := bot.Db.View(func(tx *bolt.Tx) error {
//		for {
//			price := (<-coin.Ticker)[0]["trade_price"].(float64)
//
//			balances, err := upbit.API.GetBalances(bot.Accounts)
//			if err != nil {
//				return err
//			}
//
//			if balances["KRW"] >= upbit.MinimumOrderPrice && balances["KRW"] > coin.Order && coin.Order > upbit.MinimumOrderPrice {
//				volume := coin.Order / price
//
//				if math.IsInf(volume, 0) {
//					return errors.New("division by zero")
//				}
//
//				dayCandles, err := upbit.API.GetCandlesDays("KRW-" + coin.Name, "2")
//				if err != nil {
//					return err
//				}
//
//				if coinBalance, ok := balances[coin.Name]; p.Is(dayCandles, price) {
//					if PenetrationCoins.Get([]byte("KRW-"+coin.Name)) != nil && ok {
//						// 매도
//						avgBuyPrice, err := upbit.API.GetAverageBuyPrice(bot.Accounts, coin.Name)
//						if err != nil {
//							bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
//						}
//
//						pp := price / avgBuyPrice
//
//						if math.IsInf(pp, 0) {
//							return errors.New("division by zero")
//						}
//
//						minutesCandles, err := upbit.API.GetCandlesMinutes("1", "KRW-" + coin.Name, "3")
//						if err != nil {
//							bot.Logger.Fatal(err)
//						}
//
//						// 코인이 상승했지만 하락장 일떄
//						if pp-1 >= p.H && !upbit.API.GetMarketConditionInMinutes(minutesCandles, price) {
//							p.Notify("ask", coinBalance, price)
//							continue
//						}
//						// 매도를 했지만 버킷에서 해당 종목을 제거하지 않는 이유는 판 종목을 다시 사는 것을 방지하기 위함
//
//					} else {
//						// 이미 매수했거나, 초기화 과정에서 추가된 코인이 아닌 경우에만 매수한다.
//						p.Notify("bid", volume, price)
//
//						bot.LogChan <- bot.Log{
//							Msg: "PENETRATION",
//							Fields: logrus.Fields{
//								"coin": coin.Name, "price": price, "change-rate": dayCandles[0]["change_rate"].(float64),
//							},
//							Level: logrus.InfoLevel,
//						}
//						err := PenetrationCoins.Put([]byte("KRW-"+coin.Name), []byte(fmt.Sprintf("%f", upbit.Config.C)))
//						if err != nil {
//							return err
//						}
//
//						continue
//					}
//				}
//			}
//
//			time.Sleep(1 * time.Second)
//		}
//	})
//	if err != nil {
//		bot.LogChan <- bot.Log{Msg: err.Error(), Level: logrus.FatalLevel}
//	}
//}
