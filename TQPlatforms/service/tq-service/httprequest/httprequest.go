package httprequest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"strconv"
  util "server/service/tq-service/util"
)

//Post http get method
func Get(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		util.Error(err.Error())
		return nil, errors.New("new request is fail ")
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	util.Infof("Go GET URL : %s ", req.URL.String())
	return client.Do(req)
}

//DefaultGet
func DefaultGet(url string, params map[string]string) (*http.Response, error) {
	return Get(url, params, nil)
}

//Post http post method
func Post(url string, body map[string]string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			util.Error(err.Error())
			return nil, errors.New("http post body to json failed")
		}
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))
	if err != nil {
		util.Error(err.Error())
		return nil, errors.New("new request is fail: %v \n")
	}
	req.Header.Set("Content-type", "application/json")
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	util.Infof("Go POST URL : %s  ", req.URL.String())
	return client.Do(req)
}

//Post http post method
func PostJson(url string, body map[string]interface{}, params map[string]string, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			util.Info(fmt.Sprintf("%v, body:%v", err, body))
			return nil, errors.New(fmt.Sprintf("http post body to json failed, %v", body))
		}
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))
	if err != nil {
		util.Error(err.Error())
		return nil, errors.New("new request is fail: %v \n")
	}
	req.Header.Set("Content-type", "application/json")
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	util.Infof("Go POST URL : %s ", req.URL.String())
	return client.Do(req)
}

func DefaultPost(url string, body map[string]string) (*http.Response, error) {
	return Post(url, body, nil, nil)
}

//Parse parse http response
func Parse(resp *http.Response) (interface{}, error) {
	defer resp.Body.Close()
	util.Infof("HTTP code: %d ", resp.StatusCode)
	byArr, err := ioutil.ReadAll(resp.Body)
	if bytes.ContainsAny(byArr, ">") {
		return string(byArr), nil
	}
	if err != nil {
		util.Error(err.Error())
		return "", err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, byArr, "   ", "\t")
	if err != nil {
		util.Error(err.Error())
		return "", err
	}
	return string(prettyJSON.Bytes()), err
}

func Parse2Json(resp *http.Response ) (interface{}, map[string]interface{}) {
	var data map[string]interface{}

	defer resp.Body.Close()
	util.Infof("HTTP code: %d ", resp.StatusCode)
	byArr, err := ioutil.ReadAll(resp.Body)
	if bytes.ContainsAny(byArr, ">") {
		return string(byArr), nil
	}
	if err != nil {
		util.Error(err.Error())
		return nil, data
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, byArr, "   ", "\t")
	b := prettyJSON.Bytes()
	//log.Println("b =",b)

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil,data
	}

	//log.Println("type =",reflect.TypeOf(data),data)
	return nil,data
}

func PrintResult(resp *http.Response, err error, t *testing.T) {
	if err != nil {
		t.Errorf("err = %s \n", err.Error())
		return
	}
	parse, err := Parse(resp)
	if err != nil {
		t.Errorf("err = %s \n", err.Error())
		return
	}
	util.Infof("DATA=%v ", parse)
}

func Int642String(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Float642String(i float64) string {
	return strconv.FormatFloat(i,'f', 5, 32)
}

func String2int64(str string) int64 {
	ret,_ :=  strconv.ParseInt(str, 10, 64)
	return ret
}
