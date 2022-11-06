package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func todoHandleErrorBetter(err error) {
  if err != nil {
    log.Fatalln(err)
  }
}

func checkStatusCode[T any](resp *http.Response, body T) {
  if resp.StatusCode > 299 || resp.StatusCode < 200 {
    log.Printf(
      "WARNING: status code %s on %s request on \n\turl: %v\n\trequest: %#v\n\treq data: %#v\n",
      resp.Status,
      resp.Request.Method,
      resp.Request.URL, 
      resp.Request,
      body)
  }
}

func responseToJson[R any](resp *http.Response) R {
	var rData R

	rBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Problem reading response: %#v\n\terror: %s\n", resp, err)
  }
	resp.Body.Close()

  log.Printf("Info on json response: \n\turl: %v\n\tjson: %s\n", resp.Request.URL, rBody)
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


func getRequest[R any](uri *url.URL, headers map[string][]string) R {
	req := http.Request {
		Method: "GET",
		URL: uri,
		Header: headers,
	}
	resp, err := http.DefaultClient.Do(&req)
  if err != nil {
    log.Fatalf(
      "Could not execute GET request\n\turl: %v\n\tdata: %#v\n\terror: %s\n", 
      uri, 
      req,
      err)
  }
  checkStatusCode(resp, "*GET request; no body*")

  return responseToJson[R](resp)
}
