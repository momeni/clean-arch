package repo

type Tx interface {
	Queryer
	IsTx()
}
