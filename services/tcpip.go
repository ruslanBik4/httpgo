/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package services

import (
	"bytes"
	"crypto/tls"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/ruslanBik4/logs"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

func DoPostRequest(url string, params map[string]string, hdr *fasthttp.RequestHeader) (*fasthttp.Response, error) {

	if strings.HasPrefix(url, ":") {
		url = "http://localhost" + url
	}

	req := &fasthttp.Request{}
	if hdr != nil {
		hdr.CopyTo(&req.Header)
	}

	if params != nil {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for name, val := range params {
			err := w.WriteField(name, val)
			if err != nil {
				return nil, err
			}
		}

		if err := w.Close(); err != nil {
			return nil, errors.Wrap(err, "w.Close")
		}
		req.Header.SetContentType(w.FormDataContentType())
		req.SetBody(b.Bytes())
	} else {
		req.Header.SetContentType("application/x-www-form-urlencoded")
	}

	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(url)

	c := fasthttp.Client{
		MaxIdleConnDuration: time.Minute * 10,
		MaxConnDuration:     time.Minute * 10,
		MaxConnsPerHost:     10000,
	}

	for {
		resp := &fasthttp.Response{}
		err := c.DoTimeout(req, resp, time.Minute*3)
		switch err {
		case nil:
			return resp, nil
		case fasthttp.ErrTimeout, fasthttp.ErrDialTimeout:
			logs.DebugLog("timeout %+v", resp)
			<-time.After(time.Minute * 2)
			continue
		case fasthttp.ErrNoFreeConns:
			logs.DebugLog("timeout %v", resp)
			<-time.After(time.Minute * 2)
			continue
		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				logs.DebugLog("%v %v", err, resp)
				<-time.After(time.Minute * 2)
				continue
			} else if strings.Contains(err.Error(), "socket: too many open files") {
				<-time.After(time.Second * 10)
				continue
			} else {
				return nil, err
			}
		}
	}
}

func DoPostJSONRequest(url string, json []byte, hdr *fasthttp.RequestHeader) (*fasthttp.Response, error) {

	if strings.HasPrefix(url, ":") {
		url = "http://localhost" + url
	}

	req := &fasthttp.Request{}
	if hdr != nil {
		hdr.CopyTo(&req.Header)
	}

	req.Header.SetContentType("application/json")
	req.SetBody(json)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(url)

	logs.DebugLog(req.Header.String())

	c := fasthttp.Client{
		MaxIdleConnDuration: time.Minute * 10,
		MaxConnDuration:     time.Minute * 10,
		MaxConnsPerHost:     10000,
	}

	for {
		resp := &fasthttp.Response{}
		err := c.DoTimeout(req, resp, time.Minute*3)
		switch err {
		case nil:
			return resp, nil
		case fasthttp.ErrTimeout, fasthttp.ErrDialTimeout:
			logs.DebugLog("timeout %+v", resp)
			<-time.After(time.Minute * 2)
			continue
		case fasthttp.ErrNoFreeConns:
			logs.DebugLog("timeout %v", resp)
			<-time.After(time.Minute * 2)
			continue
		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				logs.DebugLog("%v %v", err, resp)
				<-time.After(time.Minute * 2)
				continue
			} else if strings.Contains(err.Error(), "socket: too many open files") {
				<-time.After(time.Second * 10)
				continue
			} else {
				return nil, err
			}
		}
	}
}

func DoGetRequest(url string, hdr *fasthttp.RequestHeader) (*fasthttp.Response, error) {
	return doRequuest(url, hdr, fasthttp.MethodGet)
}

func DoHEADRequest(url string, hdr *fasthttp.RequestHeader) (*fasthttp.Response, error) {
	return doRequuest(url, hdr, fasthttp.MethodHead)
}

func DoRequest(url string, hdr *fasthttp.RequestHeader, method string) (*fasthttp.Response, error) {
	return doRequuest(url, hdr, method)
}

func doRequuest(url string, hdr *fasthttp.RequestHeader, method string) (*fasthttp.Response, error) {
	req := &fasthttp.Request{}

	if hdr != nil {
		hdr.CopyTo(&req.Header)
	}

	req.Header.SetMethod(method)
	req.SetRequestURI(url)

	c := fasthttp.Client{}
	//avoid error when server has not trusted certificate
	c.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	for {

		resp := &fasthttp.Response{}
		err := c.DoTimeout(req, resp, time.Minute)
		switch err {
		case nil:
			return resp, nil

		case fasthttp.ErrTimeout, fasthttp.ErrDialTimeout:
			logs.DebugLog("timeout %v", resp)
			<-time.After(time.Minute * 2)
			continue

		case fasthttp.ErrNoFreeConns:
			logs.DebugLog("timeout %v", resp)
			<-time.After(time.Minute * 2)
			continue

		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				logs.DebugLog("%v %v", err, resp)
				<-time.After(time.Minute * 2)
				continue
			} else {
				return nil, err
			}
		}
	}
}

func DoRequestAndKeep(url string, params map[string]string, ch chan<- *fasthttp.Response) error {

	if strings.HasPrefix(url, ":") {
		url = "http://localhost" + url
	}

	req := &fasthttp.Request{}

	if params != nil {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for name, val := range params {
			err := w.WriteField(name, val)
			if err != nil {
				return err
			}
		}

		if err := w.Close(); err != nil {
			return errors.Wrap(err, "w.Close")
		}
		req.Header.SetContentType(w.FormDataContentType())
		req.SetBody(b.Bytes())
	} else {
		req.Header.SetContentType("application/x-www-form-urlencoded")
	}

	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(url)

	c := fasthttp.Client{
		MaxIdleConnDuration: time.Minute * 20,
		MaxConnDuration:     time.Minute * 20,
		MaxConnsPerHost:     100000,
	}

	for {
		// todo optimize after new branch of fasthhttp
		resp := &fasthttp.Response{}
		err := c.DoTimeout(req, resp, time.Second*10)
		switch err {
		case nil:
			switch sc := resp.StatusCode(); sc {
			case fasthttp.StatusCreated, fasthttp.StatusNoContent:
				logs.DebugLog(sc)
				close(ch)
				return nil

			case fasthttp.StatusOK:
				req.Header.Set(fasthttp.HeaderConnection, "Upgrade")
				req.Header.Set("Upgrade", "cmd")

			case fasthttp.StatusPartialContent:
				// logs.DebugLog("unknow status code %d", sc)
			default:
				if sc >= 300 {
					return errors.Errorf("bad responce %d %s", sc, string(resp.Body()))
				}
			}

			ch <- resp
			continue

		case fasthttp.ErrTimeout, fasthttp.ErrDialTimeout:
			logs.DebugLog("timeout %+v", resp)
			<-time.After(time.Minute * 2)
			continue

		case fasthttp.ErrNoFreeConns:
			logs.DebugLog("no free conn %v", resp)
			<-time.After(time.Minute * 2)
			continue

		case io.EOF, io.ErrUnexpectedEOF:
			close(ch)
			logs.ErrorLog(err, resp)
			return err

		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				logs.ErrorLog(err, "%v", resp)
				<-time.After(time.Minute * 2)
				continue
			} else if strings.Contains(err.Error(), "socket: too many open files") {
				<-time.After(time.Second * 10)
				continue
			} else {
				return err
			}
		}
	}
}
