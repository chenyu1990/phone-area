package http

import (
	"errors"
	"io/ioutil"
	"net/http"
)

func Get(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}


	//connectTimer := time.NewTimer(5 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is null")
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, errors.New("request failed")
}

// 超时处理
//func doTimeoutRequest(client *http.Client, timer *time.Timer, req *http.Request) (*http.Response, error) {
//	type result struct {
//		resp *http.Response
//		err  error
//	}
//	done := make(chan result, 1)
//	go func() {
//		resp, err := client.Do(req)
//		done <- result{resp, err}
//	}()
//	select {
//	case r := <-done:
//		return r.resp, r.err
//	case <-timer.C:
//		return nil, errors.New("request timeout")
//	}
//}
