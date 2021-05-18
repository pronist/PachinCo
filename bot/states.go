package bot

// 마켓의 추적 상태를 나타내는 상수다.
// 추적이 시작된 마켓에 대해서는 tracked 상수가, 해제된 마켓에 대해서는 untracked 가 사용된다.
const (
	tracked = iota
	untracked
)

var stat = make(map[string]int)
