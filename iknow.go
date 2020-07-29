package main

import (
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
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/learning"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/config"
	"github.com/superryanguo/iknow/processor"
	"github.com/superryanguo/iknow/utils"
)

const (
	LogFile    string = "my.log"
	HoTgtModel string = "./train/train_model/HoTgt.Model"
	HoSrcModel string = "./train/train_model/HoSrc.Model"
	HoSrcTept  string = "./train/train_tempt/HoSrc.tmpt"
	HoTgtTept  string = "./train/train_tempt/HoTgt.tmpt"
)

type DataContext struct {
	Token    string
	Result   string
	Template feature.FeatureTemplate
	//TODO: should combine or split with the feature.MsgTpt
	//which way is better, should we put MsgTpt in feature package or
	//Just here?
	FeaRaw     feature.FeatureRawChain
	FeaPur     feature.FeaturePureChain
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
		var mdfile, tmptfile string
		e = r.ParseForm()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		r.ParseMultipartForm(32 << 20) //defined maximum size of file
		context.Returncode = "Form Parse done"
		formToken := template.HTMLEscapeString(r.Form.Get("CSRFToken"))
		mode := template.HTMLEscapeString(r.Form.Get("Mode"))
		model := template.HTMLEscapeString(r.Form.Get("Model"))
		context.Token = formToken
		n := strings.Split(r.RemoteAddr, ":")[0] + "-" + strings.TrimLeft(strings.Fields(r.UserAgent())[1], "(")
		uname := strings.TrimRight(n, ";")
		bodyin := template.HTMLEscapeString(r.Form.Get("bodyin"))
		var matchresult bool = false
		var fts feature.FeatureTestStatus
		cookie, e := r.Cookie("csrftoken")
		if e != nil {
			log.Warn(e)
			context.Returncode = "cookie read error" + e.Error()
			goto SHOW
		}

		if model == "HoSrc" {
			mdfile, tmptfile = HoSrcModel, HoSrcTept
		} else if model == "HoTgt" {
			mdfile, tmptfile = HoTgtModel, HoTgtTept
		} else {
			mdfile, tmptfile = "", "" //TODO: nothing
		}

		if mode == "MachineLearning" {
			context.Template, e = feature.ExtractFeatureTemplate(tmptfile)
			if e != nil {
				log.Warn(e)
				context.Returncode = "InputDataError:" + e.Error()
				//TODO: should the returncode method combine into below:
				//http.Error(w, e.Error(), http.StatusInternalServerError)
				goto SHOW
			}
		} else {
			context.Template, e = feature.ExtractFeatureTemplateHtml(bodyin)
			if e != nil {
				log.Warn(e)
				context.Returncode = "InputDataError:" + e.Error()
				//TODO: should the returncode method combine into below:
				//http.Error(w, e.Error(), http.StatusInternalServerError)
				goto SHOW
			}
		}
		feature.MsgTpt = context.Template
		feature.MsgMap.Build(feature.MsgTpt)
		feature.MsgMap.Print()
		log.Infof("%s %s %s  with cookie token %s and form token %s, Mode:%s,Model:%s\n",
			ti, uname, r.Method, cookie.Value, context.Token, mode, model)
		summary = ti + "|" + r.Method + "|" + mode + "|" + model
		log.Debug("Input :\n", bodyin)

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
				upload := "./runcmd/" + formToken + "/" + LogFile
				_, e := os.Stat(upload)
				if e == nil {
					log.Debug("upload file already exist, rm it first...")
					shell := "rm -fr " + upload
					log.Debug("run cmd ", shell)
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
				io.Copy(f, file)
				f.Close()
				context.Returncode = "upload file done"

				context.FeaRaw, e = feature.CaptureFeautres(upload)
				if e != nil {
					context.Result = e.Error()
					context.Returncode = "CaputreFeatures Fail!"
					goto SHOW
				}
				context.FeaRaw.Print()
				context.FeaPur, e = feature.TransformFeaturePure(feature.PureDuplicate(context.FeaRaw))
				if e != nil {
					context.Result = e.Error()
					context.Returncode = "TransformFeaturePure Fail!"
					goto SHOW
				}
				context.FeaPur.Print()
				//run cmd for what you want
				if mode == "TemplateMatch" {
					fts = feature.BuildTestStatus(context.FeaPur)
					log.Debug("TemplateMatch fts=", fts)
					matchresult, e = processor.TemplateMatch(fts, context.Template)
					if e != nil {
						log.Warn(e)
						context.Result = e.Error()
						context.Returncode = fmt.Sprintf("TemplateMatch Error:%s", e.Error())
					} else {
						if matchresult {
							context.Result = fmt.Sprintf("Successfully Match!\n")
							context.Returncode = "Match"
							summary += "|Done And Match"
						} else {
							context.Result = fmt.Sprintf("Log and Template MisMatch!\n")
							context.Returncode = "MisMatch"
							summary += "|Done And MisMatch"
						}
					}
				} else if mode == "MachineLearning" {
					ml, e := learning.SvmLearn(mdfile, context.FeaPur, context.Template)
					if e != nil {
						log.Warn(e)
						context.Result = e.Error()
						context.Returncode = fmt.Sprintf("TemplateMatch Error:%s", e.Error())
					} else {
						context.Result = "MachineLearning Successfully! Result=" + fmt.Sprintf("%f", ml)
						context.Returncode = "MachineLearning done!" + fmt.Sprintf("%f", ml)
						summary += "Succ" + fmt.Sprintf("%f", ml)
					}
				} else {
					log.Warn("Unknow parse mode")
					context.Returncode = "Unknown parse mode!"

				}
			} else {
				context.Returncode = "Can't read the src file!"
				log.Warn("Can't create log file, maybe nil or empty upload filename")
			}
		} else {
			log.Warn("form token mismatch")
			context.Returncode = "form token mismatch"
		}

	SHOW:
		dstore.SendData(uname, summary)
		b, e := template.ParseFiles("./templates/knowit.html")
		if e != nil {
			log.Warn(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof("Template:%v\n", context.Template)
		log.Infof("Returncode:%v\n", context.Returncode)
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
