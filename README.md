# 업비트 트레이딩 봇

업비트에서 트레이딩 봇을 구동하기 위해 만든 봇입니다. 
업비트는 어차피 한국인이 대부분 사용하므로 README 도 **한국어**로 작성합니다. 물론 소스코드의 주석도 한국어로 작성되었습니다.

<div align=center>
    <img src="https://blog.kakaocdn.net/dn/rnYj6/btq5k0aOyfk/owkkNDdBK91OLXZt48O5D0/img.gif">
</div>

## 설치

이 프로젝트를 올바르게 동작시키기 위해서는 먼저 소스코드를 다운로드 받을 필요가 있습니다. 
또한 일반적인 라이브러리나 ```go install``` 로 바로 설치하여 사용하는 프로그램이 아닌 만큼, 소스코드를 직접 다운받고 **컴파일**해야 합니다.

```bash
$ git clone https://github.com/pronist/upbit-trading-bot
```

## 시작하기

미리 작성되어 있는 예제 봇을 실행시키려면 다음과 같이 실행하십시오. 

```bash
$ go build && ./upbit
```

## 기능

해당 봇은 사용자를 위한 유용한 기능을 제공합니다.

- 전략: 자신만의 전략을 만들어보고 시험해볼 수 있습니다.
- 가상 계좌: 실제 업비트 계좌가 아닌 가상의 계좌를 사용하여 전략을 시험해볼 수 있습니다.
- 분산 투자: 투자 비율을 설정하여 분산 투자를 실현할 수 있습니다.
- 종목 탐지: 마켓을 순회하면서 조건에 맞는 종목을 탐지하여 전략을 실행시킬 수 있습니다.

봇이 동작하는 일부 원리에 대해서는 해당 프로젝트를 만든 개발자의 블로그에서
[암호화폐 트레이딩 봇을 만들었다 (feat. 업비트)](https://pronist.tistory.com/133) 를 참고해주시기 바랍니다.

## 환경설정

봇을 동작시키기 이전에 일부 환경설정에 대해 이해할 필요가 있습니다. 
환경설정은 키를 설정하거나 자금의 비율을 설정하는 등 봇을 운영함에 있어 중요한 부분을 설정할 수 있도록 합니다.
환경설정은 [config.example.yml](https://github.com/pronist/upbit-trading-bot/blob/main/config.example.yml) 에 정의합니다.

아래의 환경설정을 참고해보십시오. 접근하려면 ```static.Config``` 전역변수로 접근합니다.

```yaml
# 실제 업비트 계좌를 통해 거래를 하기 위해서는 엑세스키와 비밀키를 반드시 설정해야 한다.
#keypair:
#  accesskey: UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4
#  secretkey: 2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn

# 코인이 가질 자금 대비 최대 보유비중,
# KRW 100, 값이 0.2 일때 'A' 코인이 가질 수 있는 최대 비중은 20%
#tradablebalanceratio: 0.2

# 'A' 코인에 10 만큼 할당이 되었을 때, 0.5 이라면 50% 인 5 를 사용하여 주문
# 업비트의 최소 주문가격이 5000 KRW 이므로 주의할 것.
#orderratio: 0.5

# 추적 할 최대 마켓의 수
#maxtrackedmarket: 5

# 수동으로 추적할 마켓
#whitelist:
#  - KRW-BTC
#  - KRW-ETH
#  - KRW-BTT
#  - KRW-ADA

#blacklist:
#  - KRW-DOGE
```

**keypair**

```keypair``` 에는 ```accesskey```, ```secretkey``` 를 설정해야 합니다. 이는 [업비트 Open API](https://upbit.com/service_center/open_api_guide) 사용에서 참고할 수 있습니다.

**tradablebalanceratio**

```tradablebalanceratio``` 는 보유 비중입니다. 가지고 있는 KRW 에 대해 분산하여 투자할 때 사용합니다. 
값이 ```0.1``` 이라면 하나의 마켓에 대해 ```10%``` 를 적용하고, ```1.0``` 을 적용하면 하나에 몰빵을 말합니다. 
```최대 주문 금액 = 총 자금 * tradablebalanceratio```, 기본값은 ```0.1``` 입니다.

**orderratio**

```orderratio``` 는 주문 비중입니다. 예를 들어 KRW 100 이 있고, 
```tradablebalanceratio``` 에 ```0.1``` 이 설정되어 ```KRW-BTC``` 마켓에 ```10``` 만큼을 사용할 수 있는 경우, 
이 설정의 값이 ```0.5``` 로 설정되어 있다면 주문은 ```5``` 만큼만 가능합니다. ```주문 가격 = 최대 주문 금액 * orderratio```, 
추가적으로 업비트의 최소 주문 가격은 5000 KRW 이므로 이를 주의하십시오. 기본값은 ```0.5``` 입니다.

**maxtrackedmarket**

```maxtrackedmarket``` 는 추적할 최대 마켓의 수입니다. 
설정하지 않으면 기본 값으로 ```10``` 이 적용됩니다.

**whitelist**

```whitelist``` 가 설정되어 있다면 해당 마켓만 추적합니다. ```blacklist``` 보다 우선순위가 낮습니다. 
같은 마켓이 ```whitelist``` 와 ```blacklist``` 에 설정되어 있다면 ```blacklist``` 가 먼저 적용됩니다.

**blacklist**

```blacklist``` 가 설정되어 있다면 해당 마켓은 추적하지 않습니다. 

### 새로운 환경설정 추가하기

새로운 환경설정을 추가하려면 [static/config.go](https://github.com/pronist/upbit-trading-bot/blob/main/static/config.go) 에 
```Config``` 전역변수에 필드를 추가하면 새로운 환경변수를 얻어올 수 있습니다.

## 봇

먼저 예제로 작성된 [main.go](https://github.com/pronist/upbit-trading-bot/blob/main/main.go) 를 살펴보면 아래와 같은 코드가 있습니다. 
봇에 대한 것은 ```bot``` 패키지에 정의되어 있으며 ```bot.New``` 으로 봇을 만듭니다. 
이때, 봇에서 사용할 전략을 넘겨주는데, 
이러한 전략들은 모두 ```bot.Strategy``` 인터페이스를 만족해야 합니다.

```bot.Run``` 메서드를 사용하면 봇을 실행합니다.

```go
func main() {
	///// 봇에 사용할 전략을 설정한다.
	b := bot.New([]bot.Strategy{
		// https://wikidocs.net/21888
		&bot.PenetrationStrategy{},
	})
	/////

	///// 봇에 사용할 계정을 설정한다.
	//acc, err := bot.NewUpbitAccounts(b)
	acc, err := bot.NewFakeAccounts("accounts.db", 55000.0) // 테스트용 계정
	if err != nil {
		logrus.Fatal(err)
	}

	b.SetAccounts(acc)
	/////

	logrus.Panic(b.Run())
}
```

## 계좌

또한 봇을 만든 이후에는 봇이 사용할 계정을 설정해야 하는데, 실제 **업비트계좌**을 사용하거나 **가상계좌**를 사용할 수 있습니다.

### 가상

```NewFakeAccounts``` 를 사용하면 **가장계좌**를 만들 수 있고
첫번째 파라매터로 받는 것은 업비트의 서버에 저장하는 것 대신 로컬에 데이터베이스에 저장하도록 할 수 있습니다.

아래의 코드는 자산데이터를 ```accounts.db``` 에 저장하고 ```55000.0``` KRW 로 할당함을 의미합니다.

```go
func main() {
	acc, err := bot.NewFakeAccounts("accounts.db", 55000.0) // 테스트용 계정
	if err != nil {
		logrus.Fatal(err)
	}

	b.SetAccounts(acc)
}
```

### 업비트

```NewUpbitAccounts``` 를 통해 실제 **업비트**에서 거래를 하려고 하는 경우 아래와 같이 생성하고,

```go
func main() {
	acc, err := bot.NewUpbitAccounts(b)
	if err != nil {
		logrus.Fatal(err)
	}

	b.SetAccounts(acc)
}
```
[config.example.yml](https://github.com/pronist/upbit-trading-bot/blob/main/config.example.yml) 에 
업비트에서 발급받은 ```accesskey```, ```secretkey``` 가 설정되어 있아야 합니다. 아래의 설정은 키값의 **예**입니다.

```yaml
# 엑세스키와 비밀키를 설정해야 한다.
keypair:
  accesskey: UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4
  secretkey: 2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn
```

## 전략

전략을 만들고, 마켓을 감지하는 일은 트레이딩 봇의 핵심 요소입니다. 모든 전략은 ```bot.Strategy``` 인터페이스를 만족해야 하여 **bot** 디렉토리 아래에 위치합니다. ```bot.New``` 을 사용하여 봇을 만들때 사용됩니다.

```go
type Strategy interface {
	register(bot *Bot) error                                       // 봇이 실행될 때 전략이 최초로 등록될 때
	boot(bot *Bot, c *coin) error                                  // 코인을 생성하고 전략을 실행하기 직전
	run(bot *Bot, c *coin, t map[string]interface{}) (bool, error) // 전략
}
```

```Strategy.register``` 는 봇에 등록하고 전략을 준비시킬 때 사용합니다. *단 한번만* 실행되는 점에 유의합니다.
```Strategy.boot``` 매 전략이 시작되기 전에 호출됩니다. 디텍터가 마켓을 감지하여 전략을 실행하기 직전에 실행됩니다.
```Strategy.run``` 은 전략의 본체입니다. 여기서 실제로 주문을 하고 그 결과를 반환해야 합니다.

[bot/strategy_penetration.go](https://github.com/pronist/upbit-trading-bot/blob/main/bot/strategy_penetration.go) 에는 **변동성 돌파전략** 을 사용하도록 예시로 작성되었습니다.

### 탐지

```bot.predicate``` 함수를 작성하면 **디텍터가 감지할 마켓에 대한 조건**을 설정할 수 있습니다. 
또한 ```bot.run``` 에서는 ```detector.d``` 채널로 감지된 마켓의 틱을 얻어옵니다.

```go
for tick := range d.d {
    // 디텍팅되어 가져온 코인에 대해서 틱과 전략 시작 ...
}
```

마켓이 감지되면 ```bot.tick``` 이 같이 실행되며 이 함수는 시세정보를 **매초**마다 얻어와 ```strategy.run``` 에 넘겨줍니다.

[bot/strategy_penetration.go](https://github.com/pronist/upbit-trading-bot/blob/main/bot/strategy_penetration.go) 에는 
전략 뿐만 아니라 ```bot.predicate``` 함수도 예시로 작성되어 있습니다. 
전략을 만들기 위해서는 이러한 **전략**과 **마켓 탐지** 가 구현되어 있어야 합니다.

### 전략을 만들기 위한 팁

일반적으로 전략에 자주 사용되는 함수들은 [bot/accounts_utils.go](https://github.com/pronist/upbit-trading-bot/blob/main/bot/accounts_utils.go) 또는
[bot/utils.go](https://github.com/pronist/upbit-trading-bot/blob/main/bot/utils.go) 를 사용하고, 
```bot``` 에 있고 ```client``` 패키지에 있는  ```Client.Call```, ```QuotationClient.Call``` 을 주로 사용하여 업비트 서버에 요청합니다.

시세 정보를 얻고 싶은 때는 ```client.WebsocketClient``` 를 사용합니다. 
이 경우 [WebSocket을 이용한 업비트 시세수신](https://docs.upbit.com/docs/upbit-quotation-websocket) 을 참고하십시오.

그 외의 API 는 [업비트 Open API 레퍼런스](https://docs.upbit.com/reference) 를 참고합니다.

**주문**을 할 때는 ```Accounts.order``` 를 사용하십시오. **매수**할 때는 `b`, **매도**할 때는 `s` 상수를 ```side``` 에 넘깁니다.

## 로깅

봇에 **로그**를 남기고 싶을 때는 ```log.Logger``` 를 통해 넘길 수 있습니다. 
이는 [logrus](https://github.com/sirupsen/logrus) 로거입니다.
로그 자체는 [log/logger.go](https://github.com/pronist/upbit-trading-bot/blob/main/log/logger.go) 에서 처리됩니다.

```go
log.Logger <- log.Log{
    Msg:    "Detected",
    Fields: logrus.Fields{"market": market},
    Level:  logrus.DebugLevel,
}
```

## 상태

마켓의 추적 상태를 변경하여 전략을 실행하거나 멈추는 것을 제어합니다.
이러한 상태는 [bot/stat.go](https://github.com/pronist/upbit-trading-bot/blob/main/bot/stat.go) 에 정의되어 있습니다.

기본적으로 전략이 시작되면 해당 마켓은 ```staged``` 상태가 되며 임의로 전략 내부에서 다음과 같이 상태를 조정할 수도 있습니다.

```go
stat["KRW-BTC"] = untracked
```

모든 전략과 틱은 마켓의 상태가 ```staged``` 상태인 경우에는 동작하며 외부 요인에 의해 상태가 변하는 경우 종료됩니다.
```bot.strategy``` 에서는 다음과 같이 작동하여 상태가 변하면 반복문이 멈추게됩니다.

```go
func (b *Bot) strategy(c *coin, strategy Strategy) {
	stat, ok := stat[targetMarket+"-"+c.name]

	for ok && stat == staged {
		t := <-c.t
	}
}
```

```bot.tick``` 또한 마찬가지로 마켓의 추적 상태에 따라 멈추거나 지속적으로 실행됩니다.

```go
func (b *Bot) tick(c *coin) {
	m := targetMarket + "-" + c.name
 
	for stat[m] == staged {
		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.strategies {
			c.t <- r
		}

		time.Sleep(time.Second * 1)
	}
}
```

따라서 정리해보면, ```Bot.run``` 이 실행되면 ```detector.run``` 이 실행되고 ```bot.predicate``` 의 조건에 부합하는 마켓을 감지하여 `detector.d` 채널에 보고합니다.
그 이후 ```strategy.run```, ```bot.tick``` 이 실행된다는 것을 알 수 있습니다.

## 저작권
 
[MIT](https://github.com/pronist/upbit-trading-bot/blob/master/LICENSE)
 
Copyright 2021. [SangWoo Jeong](https://github.com/pronist). All rights reserved.
