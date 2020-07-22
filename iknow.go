package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	dstore "github.com/superryanguo/iknow/datastore"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/config"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

type DataContext struct {
	Token      string
	Data       string
	Tempalte   feature.Tempalte
	Returncode string
}

func init() {
	config.Init()
}

func KnowHandler(w http.ResponseWriter, r *http.Request) {
	var e error
	var summary string
	ti := time.Now().Format("2006-01-02 15:04:05")
	if r.Method == "GET" {
		if r.RequestURI != "/favicon.ico" {
			var context DataContext
			t, e := template.ParseFiles("./templates/knowit.html")
			if e != nil {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
			context.Token = utils.TokenCreate()
			log.Debug(ti, "the r.method ", r.Method, "create token", context.Token)
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := http.Cookie{Name: "csrftoken", Value: context.Token, Expires: expiration}
			http.SetCookie(w, &cookie)
			e = t.Execute(w, context)
			if e != nil {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
		var context DataContext
		e = r.ParseForm()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		r.ParseMultipartForm(32 << 20) //defined maximum size of file
		context.Returncode = "Parse done"
		formToken := template.HTMLEscapeString(r.Form.Get("CSRFToken"))
		mode := template.HTMLEscapeString(r.Form.Get("Mode"))
		mesgType := template.HTMLEscapeString(r.Form.Get("MessageType"))
		context.Token = formToken
		n := strings.Split(r.RemoteAddr, ":")[0] + "-" + strings.TrimLeft(strings.Fields(r.UserAgent())[1], "(")
		uname := strings.TrimRight(n, ";")
		bodyin := template.HTMLEscapeString(r.Form.Get("bodyin"))
		cookie, e := r.Cookie("csrftoken")
		if e != nil {
			log.Warn(e)
			context.Returncode = "cookie read error" + e.Error()
			goto SHOW
		}
		context.Binstr, e = CheckAndFilterDataInput(bodyin)
		if e != nil || context.Binstr == nil {
			log.Warn(e)
			context.Returncode = e.Error() + "or nil data"
			goto SHOW
		}
		log.Infof("%s %s %s  with cookie token %s and form token %s, Mode:%s,Type:%s\n",
			ti, uname, r.Method, cookie.Value, context.Token, mode, mesgType)
		summary = ti + "|" + r.Method + "|" + mode + "|" + mesgType
		log.Info("indata :\n", bodyin)
		context.Encode = hex.EncodeToString(context.Binstr)
		if formToken == cookie.Value {
			context.Returncode = "Get EqualToken done"
			file, header, e := r.FormFile("uploadfile")
			if e != nil {
				log.Warn(e)
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
			if header != nil && header.Filename != "" {
				defer file.Close()
				dir := "./runcmd/" + formToken
				_, err := os.Stat(dir)
				if err != nil {
					if os.IsNotExist(err) {
						e = os.Mkdir(dir, os.ModePerm)
						if e != nil {
							log.Warn(e)
							context.Returncode = "Can't create the dir!"
							//return //TODO: should we return or go to the html show place?
							goto SHOW
						}
					}
				}
				context.Returncode = "create the dir done"
				upload := "./runcmd/" + formToken + "/" + ProtoFile
				_, e := os.Stat(upload)
				if e == nil {
					log.Debug("upload file already exist, rm it first...")
					shell := "rm -fr " + upload
					log.Debug("run cmd", shell)
					cmd := exec.Command("sh", "-c", shell)
					_, e := cmd.CombinedOutput()
					if e != nil {
						log.Warn(e)
						context.Returncode = "Can't remove the file already exist!"
						goto SHOW
					}
				}

				f, e := os.OpenFile(upload, os.O_WRONLY|os.O_CREATE, 0666)
				if e != nil {
					log.Warn(e)
					context.Returncode = "Can't create the file!"
					goto SHOW
				}
				defer f.Close()
				io.Copy(f, file)
				context.Returncode = "upload file done"

				//run cmd for what you want
				if mode == "Normal" {
					if mesgType != "" {
						output, e := ParseGpbNormalMode(context.Binstr, mesgType, upload)
						if e != nil {
							log.Warn(e)
							context.Decode = e.Error()
							context.Returncode = fmt.Sprintf("ParseNormalMode Error:%s", e.Error())
						} else {
							context.Decode = fmt.Sprintf("%s", output)
							context.Returncode = "Successfully Parse Normal mode done!"
							summary += "Succ"
						}
					} else {
						context.Returncode = "Error! NormalMode Must fill the messagetype"
					}
				} else if mode == "HardCore" {
					output, e := HardcoreDecode(upload, context.Binstr)
					if e != nil {
						log.Warn(e)
						context.Decode = e.Error()
						context.Returncode = fmt.Sprintf("HardCoreMode Error:%s", e.Error())
					} else {
						context.Decode = fmt.Sprintf("%s", output)
						context.Returncode = "Successfully Parse HardCore mode done!"
						summary += "Succ"

					}
				} else {
					log.Warn("Unknow parse mode")
					context.Returncode = "Unknown parse mode!"
				}

			} else {
				context.Returncode = "Can't read the src file!"
				log.Warn("Can't create the data source file, maybe nil or empty upload filename")
			}
		} else {
			log.Warn("form token mismatch")
			context.Returncode = "form token mismatch"
		}
	SHOW:
		dstore.SendData(uname, summary+"|"+context.Returncode)
		b, e := template.ParseFiles("./templates/datapost.html")
		if e != nil {
			log.Warn(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof("Encode:\n%s", context.Encode)
		log.Infof("Decode:\n%s", context.Decode)
		log.Infof("Returncode:\n%s", context.Returncode)
		e = b.Execute(w, context)
		if e != nil {
			log.Warn(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		//http.Redirect(w, r, "/", 302)
	} else {
		log.Warn("Unknown request")
		http.Redirect(w, r, "/", 302)
	}

}

func main() {
	port := flag.String("port", "8091", "Server Port")
	flag.Parse()

	go dstore.Run()
	http.HandleFunc("/", KnowHandler)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates"))))
	http.Handle("/runcmd/", http.StripPrefix("/runcmd/", http.FileServer(http.Dir("./runcmd"))))
	log.Infof("Running the server on port %q.", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
