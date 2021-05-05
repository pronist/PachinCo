package bot

type Strategy interface {
	Run(coin *Coin)
}
