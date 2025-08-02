package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	wl_http "github.com/wsva/lib_go/http"
	wl_json "github.com/wsva/lib_go/json"
	wl_net "github.com/wsva/lib_go/net"
	wl_int "github.com/wsva/lib_go_integration"
)

// 不支持指定ip，但也不用验证登录token
// https://1.1.1.1:83/get?type=ogg&type=weblogic
func handleGet(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == "GET" {
		query := r.URL.Query()
		typeList := query["type"]
		queryIP := wl_net.GetIPFromRequest(r).String()
		configList, err := GetConfigFileList(typeList, queryIP)
		if err != nil {
			wl_http.RespondError(w, err)
			return
		}
		jsonBytes, _ := json.Marshal(configList)
		fmt.Fprint(w, wl_json.Unescape(wl_json.IndentString(string(jsonBytes))))
	}
}

// 支持指定ip，但需要验证登录token
// https://1.1.1.1:83/get?ip=1.1.1.1
func handleGetByIP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == "GET" {
		query := r.URL.Query()
		_, found := query["ip"]
		if !found {
			wl_http.RespondError(w, "需传入ip")
			return
		}
		configList, err := GetConfigFileList(nil, query["ip"][0])
		if err != nil {
			wl_http.RespondError(w, err)
			return
		}
		resp := wl_http.Response{
			Success: true,
			Data: wl_http.ResponseData{
				List: configList,
			},
		}
		resp.DoResponse(w)
	}
}

func handleCheckToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !wl_int.CheckInternalKey(r, AESKey, AESIV) {
		token, err := wl_int.ParseTokenFromRequest(r)
		if err != nil {
			fmt.Println("parse token error: ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err = wl_int.CheckAndRefreshToken(cc.AccountAddress, CACrtFile, token)
		if err != nil {
			fmt.Println("check token error: ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	next(w, r)
}
