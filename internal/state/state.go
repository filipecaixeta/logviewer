package state

type State uint8

var (
	StateLoading     = State(0)
	StateBrose       = State(1)
	StateLogsLoading = State(2)
	StateLogs        = State(3)
	StateNewView     = State(3)
	StateLoadView    = State(4)
)
