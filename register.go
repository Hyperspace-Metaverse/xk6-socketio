package socketio

import "go.k6.io/k6/js/modules"

func init() {
	modules.Register("k6/x/socketio", &RootModule{moduleName: "k6/x/socketio"})
}
