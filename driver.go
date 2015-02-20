package graphie

type driverSetupFn func(*Graph) (Storage, error)

var (
	drivers = make(map[string]driverSetupFn)
)

func RegisterDriver(name string, fn driverSetupFn) {
	drivers[name] = fn
}
