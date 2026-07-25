package internal

var (
	module   string
	appName  string
	version  string
	revision string
)

func Module() string {
	return module
}

func AppName() string {
	return appName
}

func Version() string {
	return version
}

func Revision() string {
	return revision
}
