package main

import (
	"fmt"
	"github.com/m2q/aema/csgo"
)

func main() {

	hltv := csgo.HLTV{}

	err := hltv.Fetch()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	_, _ = hltv.GetFutureMatches()
	_, _ = hltv.GetPastMatches()


}
