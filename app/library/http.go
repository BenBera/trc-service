package library

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	once      sync.Once
	netClient *http.Client
)

func newNetClient() *http.Client {

	keepAliveTimeout := 120 * time.Second
	timeout := 120 * time.Second
	once.Do(func() {
		var netTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: keepAliveTimeout,
			}).Dial,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			TLSHandshakeTimeout: timeout,
		}
		netClient = &http.Client{
			Timeout:   timeout,
			Transport: netTransport,
		}
	})

	return netClient
}
func HTTPGetWithHeaders(remoteURL string, headers map[string]string, payload map[string]string) (string, int, error) {

	//t1 := GetTime()

	var fields []string

	if payload != nil {

		for key, value := range payload {

			val := fmt.Sprintf("%s=%v", key, url.QueryEscape(value))

			fields = append(fields, val)
		}
	}

	if len(fields) > 0 {

		params := strings.Join(fields, "&")
		remoteURL = fmt.Sprintf("%s?%s", remoteURL, params)

	}

	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 500, err
	}

	if headers != nil {

		for key, value := range headers {

			req.Header.Set(key, value)
		}
	}

	resp, err := newNetClient().Do(req)
	if resp != nil {

		defer resp.Body.Close()
	}

	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 500, err
	}

	st := resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 500, err
	}
	/*
		t2 := GetTime()

		if st > 299 || st < 200 {

			var fields []string

			for k, v := range headers {

				fields = append(fields, fmt.Sprintf("%s : %s", k, v))
			}

			prts := strings.Join(fields, "\n")
			log.Printf("===============START HTTP GET ===============\n\nURL: %s\n-------------------------\nHEADERS\n-------------------------\n%s\n-------------------------\nResponse HTTP Status\n-------------------------\n%d\n-------------------------\nResponse Body\n-------------------------\n%s\n-------------------------\nTime Taken\n-------------------------\n%dms\n\n===============END HTTP POST ===============", remoteURL, prts, st, string(body), t2-t1)
		}
	*/

	return string(body), st, nil
}
func HTTPPost(url string, headers map[string]string, payload interface{}) (int, string) {

	t1 := GetTime()

	jsonData, err := json.Marshal(payload)

	if err != nil {

		log.Printf("Got error decoding payload error %v ", err.Error())
		return 0, ""
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return 0, ""
	}

	var headerStrings []string

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	headerStrings = append(headerStrings, "Content-Type : application/json")
	headerStrings = append(headerStrings, "Accept : application/json")

	if headers != nil {

		for key, value := range headers {

			req.Header.Set(key, value)
			headerStrings = append(headerStrings, fmt.Sprintf("%s : %s", key, value))

		}
	}

	resp, err := newNetClient().Do(req)
	if resp != nil {

		defer resp.Body.Close()
	}

	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return 0, ""
	}

	st := resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return st, ""
	}

	//if st > 299 || st < 200 {

	log.Printf(fmt.Sprintf("===============START HTTP POST ===============\n\nURL: %s\n-------------------------\nHEADERS\n%s\n-------------------------\nBODY\n-------------------------\n%s\n-------------------------\nResponse HTTP Status\n-------------------------\n%d\n-------------------------\nResponse Body\n-------------------------\n%s\n-------------------------\nTime Taken\n-------------------------\n%dms\n\n===============END HTTP POST ===============", url, strings.Join(headerStrings, "\n"), jsonData, st, string(body), GetTime()-t1))
	//}

	return st, string(body)
}
func HTTPPost1(url string, headers map[string]string, payload interface{}) (int, string) {

	t1 := GetTime()

	jsonData, err := json.Marshal(payload)

	if err != nil {

		log.Printf("Got error decoding payload error %v ", aurora.Red(err.Error()))
		return 0, ""
	}

	jsonDatas := strings.ReplaceAll(string(jsonData), "{", "")
	//jsonDatas = strings.ReplaceAll(jsonDatas,"}","---")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonDatas)))
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return 0, ""
	}

	var headerStrings []string

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	headerStrings = append(headerStrings, "Content-Type : application/json")
	headerStrings = append(headerStrings, "Accept : application/json")

	if headers != nil {

		for key, value := range headers {

			req.Header.Set(key, value)
			headerStrings = append(headerStrings, fmt.Sprintf("%s : %s", key, value))

		}
	}

	resp, err := newNetClient().Do(req)
	if resp != nil {

		defer resp.Body.Close()
	}

	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return 0, ""
	}

	st := resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return st, ""
	}

	//if st > 299 || st < 200 {

	log.Printf("===============START HTTP POST ===============\n\nURL: %s\n-------------------------\nHEADERS\n%s\n-------------------------\nBODY\n-------------------------\n%s\n-------------------------\nResponse HTTP Status\n-------------------------\n%d\n-------------------------\nResponse Body\n-------------------------\n%s\n-------------------------\nTime Taken\n-------------------------\n%dms\n\n===============END HTTP POST ===============", url, strings.Join(headerStrings, "\n"), jsonDatas, st, string(body), GetTime()-t1)
	//}

	return st, string(body)
}
