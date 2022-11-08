package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
  "os"
)

func todoHandleErrorBetter(err error) {
  if err != nil {
    log.Fatalln(err)
  }
}

func getConfigFilePath(name string) string {
  return name // TODO get path to config directory
}


func checkStatusCode[T any](resp *http.Response, body T) {
  if resp.StatusCode > 299 || resp.StatusCode < 200 {
    text, _ := io.ReadAll(resp.Body)
    resp.Body.Close()

    log.Printf(
      "WARNING: status code %s on %s request on \n\turl: %v\n\trequest: %#v\n\treq data: %#v\n\tresp data: %#v\n",
      resp.Status,
      resp.Request.Method,
      resp.Request.URL, 
      resp.Request,
      body,
      string(text[:]))
  }
}

func logRequest(req http.Request) {
  log.Printf("Info: %s request on url (%s)", req.Method, req.URL)
}

func responseToJson[R any](resp *http.Response) R {
	var rData R

	rBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Problem reading response: %#v\n\terror: %s\n", resp, err)
  }
	resp.Body.Close()

	err = json.Unmarshal(rBody, &rData)
  if err != nil {
    log.Fatalf(
      "Problem parsing response body into type %T:\n\tbody: %s\n\terror: %s\n", 
      rData, 
      rBody, 
      err)
  }
  
  return rData
}



func postRequestNoMarshal[Q any](uri *url.URL, headers map[string][]string, data Q) *http.Response {
	qBody, err := json.Marshal(data)
  if err != nil {
    log.Fatalf(
      "Could not marshal data from type %T into json.\n\tdata: %#v\n\terror: %s\n", 
      data, 
      data,
      err)
  }
	
	req := http.Request {
		Method: "POST",
		URL: uri,
		Header: headers,
		Body: io.NopCloser(bytes.NewBuffer(qBody)),
	}
  logRequest(req)
	resp, err := http.DefaultClient.Do(&req)
  if err != nil {
    log.Fatalf(
      "Could not execute POST request on\n\turl: %v\n\tdata: %#v\n\terror: %s\n",
      uri, 
      req,
      err)
  }
  checkStatusCode(resp, data)
  
  return resp
}

func postRequest[Q any, R any](uri *url.URL, headers map[string][]string, data Q) R {
  resp := postRequestNoMarshal(uri, headers, data)
  return responseToJson[R](resp)
}


func getRequestNoMarshal(uri *url.URL, headers map[string][]string) *http.Response {
	req := http.Request {
		Method: "GET",
		URL: uri,
		Header: headers,
	}
  logRequest(req)
	resp, err := http.DefaultClient.Do(&req)
  if err != nil {
    log.Fatalf(
      "Could not execute GET request\n\turl: %v\n\tdata: %#v\n\terror: %s\n", 
      uri, 
      req,
      err)
  }
  checkStatusCode(resp, "*GET request; no body*")

  return resp
}

func getRequest[R any](uri *url.URL, headers map[string][]string) R {
  resp := getRequestNoMarshal(uri, headers)
  return responseToJson[R](resp)
}

func saveAsJsonToFile[T any](data T, filename string) {
  filename = getConfigFilePath(filename)
  datajson, err := json.Marshal(data)
  todoHandleErrorBetter(err)
  file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, 0600)
  todoHandleErrorBetter(err)
  _, err = file.Write(datajson)
  todoHandleErrorBetter(err)
  file.Close()
}

func tryLoadFromJsonToFile[T any](filename string) (T, bool) {
  var ret T
  filename = getConfigFilePath(filename)
  file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
  if err != nil {
    return ret, false
  }
  data, err := io.ReadAll(file)
  file.Close()
  json.Unmarshal(data, &ret)
  todoHandleErrorBetter(err)
  return ret, true
}
