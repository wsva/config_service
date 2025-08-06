package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"

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

func handleOAuth2Login(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	oa := oaMap.Add(&mainConfig.AuthService, httpsClient)
	oa.HandleLogin(w, r)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	state := r.FormValue("state")
	oa, err := oaMap.Get(state)
	if err != nil {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	oaMap.Delete(state)
	oa.HandleCallback(w, r)
}

func handleLogin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) == nil {
		return_to := r.FormValue("return_to")
		if return_to == "" {
			return_to = "/"
		}
		http.Redirect(w, r, return_to, http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, filepath.Join(Basepath, "template/html/login.html"))
}

func handleDashboard(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) != nil {
		thisURL := wl_net.GetFullURL(r)
		http.Redirect(w, r, fmt.Sprintf("/login?return_to=%v", url.PathEscape(thisURL)), http.StatusSeeOther)
		return
	}

	userinfoCookie, err := r.Cookie("userinfo")
	if err != nil {
		wl_http.RespondError(w, "missing user info")
		return
	}
	userinfoBytes, err := base64.URLEncoding.DecodeString(userinfoCookie.Value)
	if err != nil {
		wl_http.RespondError(w, "invalid user info")
		return
	}

	var userinfo wl_int.UserInfo
	err = json.Unmarshal(userinfoBytes, &userinfo)
	if err != nil {
		wl_http.RespondError(w, "invalid user info")
		return
	}

	tpFile := filepath.Join(Basepath, "template/html/dashboard.html")
	tp, err := template.ParseFiles(tpFile)
	if err != nil {
		fmt.Fprintf(w, "parse template %v error: %v", tpFile, err)
		return
	}
	c := struct {
		Name  string
		Email string
	}{
		Name:  userinfo.Name,
		Email: userinfo.Email,
	}
	tp.Execute(w, c)
}

func handleCheckToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !wl_int.CheckInternalKey(r, AESKey, AESIV) {
		if err := wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL); err != nil {
			fmt.Println("verify token error: ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	next(w, r)
}
