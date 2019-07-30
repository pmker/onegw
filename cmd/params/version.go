package params

import "github.com/pmker/onegw"

var Version = func() string {
	return onegw.ONEGW_BUILD_VERSION
}()
