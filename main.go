package main

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	wl_compress "github.com/wsva/lib_go/compress"
	wl_int "github.com/wsva/lib_go_integration"
)

func main() {
	err := initGlobals()
	if err != nil {
		fmt.Println(err)
		return
	}

	go crontabBackground()

	routerHttp := mux.NewRouter()

	routerHttp.PathPrefix("/css/").Handler(http.StripPrefix("/css/",
		http.FileServer(http.Dir(filepath.Join(Basepath, "template/css/")))))
	routerHttp.PathPrefix("/js/").Handler(http.StripPrefix("/js/",
		http.FileServer(http.Dir(filepath.Join(Basepath, "template/js/")))))

	routerHttp.Handle("/get",
		negroni.New(
			negroni.HandlerFunc(handleGet),
		))

	serverHttp := negroni.New(negroni.NewRecovery())
	serverHttp.Use(negroni.NewLogger())
	serverHttp.UseHandler(routerHttp)

	routerHttps := mux.NewRouter()

	routerHttps.PathPrefix("/css/").Handler(http.StripPrefix("/css/",
		http.FileServer(http.Dir(filepath.Join(Basepath, "template/css/")))))
	routerHttps.PathPrefix("/js/").Handler(http.StripPrefix("/js/",
		http.FileServer(http.Dir(filepath.Join(Basepath, "template/js/")))))

	routerHttps.Handle("/get",
		negroni.New(
			negroni.HandlerFunc(handleGet),
		))
	routerHttps.Handle("/getbyip",
		negroni.New(
			negroni.HandlerFunc(handleCheckToken),
			negroni.HandlerFunc(handleGetByIP),
		))

	routerHttps.Handle("/",
		negroni.New(
			negroni.HandlerFunc(handleDashboard),
		))
	routerHttps.Handle("/login",
		negroni.New(
			negroni.HandlerFunc(handleLogin),
		))
	routerHttps.Handle(wl_int.OAuth2LoginPath,
		negroni.New(
			negroni.HandlerFunc(handleOAuth2Login),
		))
	routerHttps.Handle(wl_int.OAuth2CallbackPath,
		negroni.New(
			negroni.HandlerFunc(handleOAuth2Callback),
		))

	serverHttps := negroni.New(negroni.NewRecovery())
	//server.Use(bha.NewCORSHandler(nil, nil, nil))
	serverHttps.Use(negroni.NewLogger())
	serverHttps.UseHandler(routerHttps)

	/*
		为了能够支持python、curl、wget访问，部分服务支持http
	*/
	for _, v := range mainConfig.ListenList {
		if !v.Enable {
			continue
		}
		v1 := v
		switch v1.LowercaseProtocol() {
		case "http":
			go func() {
				err = http.ListenAndServe(fmt.Sprintf(":%v", v1.Port),
					serverHttp)
				if err != nil {
					fmt.Println(err)
				}
			}()
		case "https":
			go func() {
				s := &http.Server{
					Addr:    fmt.Sprintf(":%v", v1.Port),
					Handler: serverHttps,
				}
				s.SetKeepAlivesEnabled(false)
				err = s.ListenAndServeTLS(ServerCrtFile, ServerKeyFile)
				if err != nil {
					fmt.Println(err)
				}
			}()
		}
	}
	select {}
}

func crontabBackground() {
	for {
		switch mainConfig.Role {
		case RoleRoot:
			err := GenerateConfigDataRoot(mainConfig.SourceList)
			if err != nil {
				fmt.Println(err)
				break
			}

			err = wl_compress.ZipCompressPath(DataDir, RootDir, ConfigDataZipFile)
			if err != nil {
				fmt.Println(err)
				break
			}

			for _, v := range mainConfig.SendToList {
				if !v.Enable {
					continue
				}
				err = v.Upload(path.Join(RootDir, ConfigDataZipFile))
				if err != nil {
					fmt.Println(err)
				}
			}

		case RoleBranch:
			err := GenerateConfigDataBranch(mainConfig.SourceList)
			if err != nil {
				fmt.Println(err)
			}

		default:
			fmt.Println("not supported role")
		}
		time.Sleep(1 * time.Minute)
	}
}
