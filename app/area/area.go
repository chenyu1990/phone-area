package area

import (
	"encoding/json"
	"fmt"
	"phone-area/schema"
	"phone-area/util/http"
)

const API = "https://cx.shouji.360.cn/phonearea.php?number="

type Data struct {
	Province        string `json:"province"`
	City            string `json:"city"`
	ServiceProvider string `json:"sp"`
}

type Response struct {
	Code int
	Data Data
}

func getInfo(info *schema.PhoneInfo, tryCounts ...int64) (*Response, error) {
	body, err := http.Get(API + info.Number)
	if err != nil {
		return nil, err
	}
	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Data.Province == "" {
		var tryCount int64
		if len(tryCounts) > 0 {
			tryCount = tryCounts[0]
		}
		if tryCount >= 10 {
			return &res, nil
		}
		tryCount++
		fmt.Printf("try %s %d", info.Number, tryCount)
		return getInfo(info, tryCount)
	}

	return &res, nil
}