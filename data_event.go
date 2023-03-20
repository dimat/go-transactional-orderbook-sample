package matchingengine

type dataEvent interface {
	Execute(*engine)
	Revert(*engine)
}
