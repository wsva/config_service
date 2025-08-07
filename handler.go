package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	wl_crypto "github.com/wsva/lib_go/crypto"
	wl_http "github.com/wsva/lib_go/http"
	wl_json "github.com/wsva/lib_go/json"
	wl_net "github.com/wsva/lib_go/net"
	wl_int "github.com/wsva/lib_go_integration"
)

// 不支持指定ip，但也不用验证登录token
// https://1.1.1.1:83/get?types=ogg,weblogic
func handleConfig(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	typeList := strings.Split(r.FormValue("types"), ",")
	ip := wl_net.GetIPFromRequest(r).String()
	configList, err := GetConfigFileList(typeList, ip)
	if err != nil {
		fmt.Println(err)
		wl_http.RespondError(w, "internal error")
		return
	}
	jsonBytes, _ := json.Marshal(configList)
	fmt.Fprint(w, wl_json.Unescape(wl_json.IndentString(string(jsonBytes))))
}

// 支持指定ip，但需要验证登录token
// https://1.1.1.1:83/get?ip=1.1.1.1
func handleConfigIP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) != nil {
		wl_http.RespondError(w, "unauthorized")
		return
	}
	ip := r.FormValue("ip")
	if ip == "" {
		wl_http.RespondError(w, "missing ip")
		return
	}
	typeList := strings.Split(r.FormValue("types"), ",")
	configList, err := GetConfigFileList(typeList, ip)
	if err != nil {
		fmt.Println(err)
		wl_http.RespondError(w, "internal error")
		return
	}
	wl_http.RespondJSON(w, wl_http.Response{
		Success: true,
		Data: wl_http.ResponseData{
			List: configList,
		},
	})
}

func handleConfigType(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) != nil {
		wl_http.RespondError(w, "unauthorized")
		return
	}
	configType := r.FormValue("type")
	if configType == "" {
		wl_http.RespondError(w, "missing type")
		return
	}
	fileinfoList, err := os.ReadDir(path.Join(DataDir, configType))
	if err != nil {
		fmt.Println(err)
		wl_http.RespondError(w, "internal error")
		return
	}
	result := make([]ConfigFile, len(fileinfoList))
	for k, v := range fileinfoList {
		result[k].Filename = v.Name()
		contentBytes, err := os.ReadFile(path.Join(DataDir, configType, v.Name()))
		if err != nil {
			fmt.Println(err)
			wl_http.RespondError(w, "internal error")
			return
		}
		decryptedContent, err := wl_crypto.AES256Decrypt(AESKey, AESIV, string(contentBytes))
		if err != nil {
			fmt.Println(err)
			wl_http.RespondError(w, "internal error")
			return
		}
		result[k].Content = decryptedContent
	}
	wl_http.RespondJSON(w, wl_http.Response{
		Success: true,
		Data: wl_http.ResponseData{
			List: result,
		},
	})
}

func handleTypeAll(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) != nil {
		wl_http.RespondError(w, "unauthorized")
		return
	}
	indexfile := path.Join(DataDir, ConfigIndexFile)
	indexlist, err := LoadConfigIndexListFromFile(indexfile)
	if err != nil {
		fmt.Println(err)
		wl_http.RespondError(w, "internal error")
		return
	}
	typeList := make([]string, len(indexlist))
	for k, v := range indexlist {
		typeList[k] = v.DirectoryOnServer
	}
	wl_http.RespondJSON(w, wl_http.Response{
		Success: true,
		Data: wl_http.ResponseData{
			List: typeList,
		},
	})
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

func handleDashboard(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	tpFile := filepath.Join(Basepath, "template/html/dashboard.html")
	tp, err := template.ParseFiles(tpFile)
	if err != nil {
		fmt.Fprintf(w, "parse template %v error: %v", tpFile, err)
		return
	}

	type Data struct {
		Name  string
		Email string
	}

	if wl_int.VerifyToken(r, httpsClient, mainConfig.AuthService.IntrospectURL) != nil {
		tp.Execute(w, Data{})
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

	tp.Execute(w, Data{
		Name:  userinfo.Name,
		Email: userinfo.Email,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	wl_int.DeleteCookieToken(w, "access_token")
	wl_int.DeleteCookieToken(w, "refresh_token")
	wl_int.DeleteCookieToken(w, "userinfo")
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
