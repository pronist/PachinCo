package bot

type Strategy interface {
	Tracking(map[string]float64, string)
}
