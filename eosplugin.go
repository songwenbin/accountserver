package accountserver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/deadbeef/eosapi"
)

type EOSActionJson struct {
	Action []EOSActionTraceJson `json:"actions"`
}

type EOSActionTraceJson struct {
	ActionTrace EOSActJson `json:"action_trace"`
}

type EOSActJson struct {
	Act EOSDataJson `json:"act"`
}

type EOSDataJson struct {
	Data EOSDataStructJson `json:"data"`
}

type EOSDataStructJson struct {
	From     string `json:from"`
	To       string `json:"to"`
	Quantity string `json:"quantity"`
	Memo     string `json:"memo"`
}

var duck map[string]float64 = make(map[string]float64)

func GetEOS() map[string]float64 {
	params := map[string]interface{}{

		"account_name": "deadbeefduck",
		"pos":          0,
		"offset":       100,
	}

	data, _ := eosapi.Post("https://nodes.get-scatter.com/v1/history/get_actions", params, nil)
	//fmt.Println(string(data))

	var m EOSActionJson
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Println("无法序列化Action")
		fmt.Println(err)
		return nil
	}

	for _, v := range m.Action {
		if v.ActionTrace.Act.Data.To == "deadbeefduck" {
			if eosvalue, exists := duck[v.ActionTrace.Act.Data.Memo]; exists == true {
				var t float64 = GetEOSValue(v.ActionTrace.Act.Data.Quantity)
				duck[v.ActionTrace.Act.Data.Memo] = eosvalue + t
			} else {
				duck[v.ActionTrace.Act.Data.Memo] = GetEOSValue(v.ActionTrace.Act.Data.Quantity)
			}
			//fmt.Println(v.ActionTrace.Act.Data)
		}
	}

	fmt.Println(duck)

	return duck
}

func GetEOSValue(EOS string) float64 {
	s := strings.Split(EOS, " ")
	var eos float64
	value, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		eos = 0.0
	} else {
		eos = value
	}

	return eos
}

type EOSPayPlugin struct {
}

func (ep EOSPayPlugin) Monitor(monitor int, chaneos chan map[string]float64) {
	var count = 0
	for {
		select {
		case <-time.After(time.Duration(monitor) * time.Second):
			count++
			if count == 5 {
				return
			}
			chaneos <- GetEOS()
		}
	}
}

func (ep EOSPayPlugin) Coin2Day(i float64) int {
	var totaltime float64 = i * 24
	return int(totaltime)
}
