package bot

// 추적 상태를 나타내는 상수들
const (
	TRACKING = iota
	STOPPED
)

var MarketTrackingStates = make(map[string]int)
