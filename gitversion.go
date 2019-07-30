package onegw

import (
	"fmt"
)

func PrintBuildVersion() {
	if  ONEGW_VERSION != "" {
		fmt.Println("this onegw node`s git GO version is ", ONEGW_VERSION)
	} else {
		fmt.Println("can not read gitversion file please use Make to build Onegw ")
	}
}
