package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/go-bongo/bongo"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	bson2 "gopkg.in/mgo.v2/bson"
)

var connection *bongo.Connection

var templates = template.Must(template.ParseFiles("templates/scam.html", "templates/manage.html", "templates/new.html"))

func main() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		FullTimestamp:             true,
	})
	flag.String("mongodb", "localhost", "MongoDB Connection String")
	flag.Int("port", 8000, "Port where the url shortener listens")
	flag.String("admin-password", "foobar2342", "Password for the admin endpoint")
	flag.String("base-url", "http://localhost:8000", "Baseurl of the URL shortener")
	_ = viper.BindPFlags(flag.CommandLine)
	flag.Parse()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	config := &bongo.Config{
		ConnectionString: viper.GetString("mongodb"),
	}
	var err error
	connection, err = bongo.Connect(config)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Connected to database")
	_ = connection.Collection("links").Collection().EnsureIndex(mgo.Index{
		Key:              []string{"name"},
		Unique:           true,
	})

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")
	r.HandleFunc("/", newShortUrl).Methods("POST")
	r.HandleFunc("/delete", deleteHandler).Methods("GET")
	r.HandleFunc("/scam", scamHandler).Methods("POST")
	r.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/admin.html")
	}).Methods("GET")
	r.PathPrefix("/favicon.ico").Handler(http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/monitoring", monitoringHandler).Methods("HEAD")
	r.HandleFunc("/monitoring", monitoringHandler).Methods("GET")
	r.HandleFunc("/{name}", redirectHandler).Methods("GET")
	r.HandleFunc("/{name}/{password}", manageHandler).Methods("GET")
	r.HandleFunc("/{name}/{password}/delete", deleteHandler).Methods("GET")
	r.HandleFunc("/{name}", redirectHandler).Methods("POST")
	r.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.PathPrefix("/").HandlerFunc(notFoundHandler)
	http.Handle("/", r)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("port")), nil))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	requestTimer := time.Now()
	if err := r.ParseForm(); err != nil {
		returnError500(err, w);
		return
	}
	params := mux.Vars(r)
	name := params["name"]
	if name == "" {
		_, _ = fmt.Fprintf(w, "We need a link name..")
		return
	}
	var link Link
	err := connection.Collection("links").FindOne(bson2.M{"name": name}, &link)
	if _, ok := err.(*bongo.DocumentNotFoundError); ok {
		logrus.Info(fmt.Sprintf("Short \"%s\" not found", name))
		returnError404(w)
		return
		// crappingfuckfuckers warum gibt es denn nen string zurück mit not found.. wie dumm.
	} else if err != nil {
		if err.Error() == "not found" {
			logrus.Info(fmt.Sprintf("Short \"%s\" not found", name))
			returnError404(w)
			return
		}
		returnError500(err, w)
		return
	}
	if params["password"] == "" || (params["password"] != viper.Get("admin-password") && params["password"] != link.Password) {
		logrus.Info(fmt.Sprintf("Auth failed on delete for \"%s\"... tried \"%s\"", name, params["password"]))
		returnError401(w)
		return
	}
	if params["password"] != viper.Get("admin-password") && link.Scam {
		logrus.Info(fmt.Sprintf("Tried to delete \"%s\" but marked as scam", name))
		_, _ = fmt.Fprint(w, "This link was marked as scam, it's disabled.")
		return
	}
	if connection.Collection("links").DeleteOne(bson2.M{"name": link.Name}) != nil {
		returnError500(err, w)
		return
	}
	requestTime := time.Since(requestTimer)
	logrus.Info(fmt.Sprintf("[%v] Deleting \"%s\"", requestTime, name))
	_, _ = fmt.Fprintf(w, "Link deleted!")
}

func scamHandler(w http.ResponseWriter, r *http.Request) {
	requestTimer := time.Now()
	if err := r.ParseForm(); err != nil {
		returnError500(err, w);
		return
	}
	name := r.FormValue("name")
	password := r.FormValue("password")
	if password != viper.Get("admin-password") {
		logrus.Info(fmt.Sprintf("Auth failed on delete for \"%s\"... tried \"%s\"", name, password))
		returnError401(w)
		return
	}
	if name == "" {
		_, _ = fmt.Fprintf(w, "We need a link name..")
		return
	}
	err, link := getLink(name)
	if err != nil {
		if err.Error() == "404" {
			returnError404(w)
			return
		}
		returnError500(err, w)
		return
	}
	link.Scam = true
	err = connection.Collection("links").Save(link)
	if err != nil {
		returnError500(err, w)
	} else {
		requestTime := time.Since(requestTimer)
		logrus.Info(fmt.Sprintf("[%v] Scamming \"%s\"", requestTime, name))
		_, _ = fmt.Fprintf(w, "Link Scammed!")
	}
}
func manageHandler(w http.ResponseWriter, r *http.Request) {
	requestTimer := time.Now()
	if err := r.ParseForm(); err != nil {
		returnError500(err, w)
		return
	}
	params := mux.Vars(r)
	name := params["name"]

	if name == "" {
		_, _ = fmt.Fprintf(w, "We need a link name.")
		return
	}
	err, link := getLink(name)
	if err != nil {
		if err.Error() == "404" {
			returnError404(w)
			return
		}
		returnError500(err, w)
		return
	}
	if params["password"] == "" || (params["password"] != viper.Get("admin-password") && params["password"] != link.Password) {
		logrus.Info(fmt.Sprintf("Auth failed on manage for \"%s\"... tried \"%s\"", name, params["password"]))
		returnError401(w)
		return
	}
	if params["password"] != viper.Get("admin-password") && link.Scam {
		_, _ = fmt.Fprint(w, "This link was marked as scam, it's disabled.")
	}
	err = templates.ExecuteTemplate(w, "manage.html", map[string]interface{}{
		"BaseUrl": viper.GetString("base-url"),
		"Link": link,
	})
	if err != nil {
		logrus.Error(err)
	}
	requestTime := time.Since(requestTimer)
	logrus.Info(fmt.Sprintf("[%v] Getting counts for \"%s\"", requestTime, name))
}
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	requestTimer := time.Now()
	params := mux.Vars(r)
	err, link := getLink(params["name"])
	if err != nil {
		if err.Error() == "404" {
			returnError404(w)
			return
		}
		returnError500(err, w)
		return
	}
	if link.Scam && r.Method != "POST" {
		err := templates.ExecuteTemplate(w, "scam.html", link)
		if err != nil {
			logrus.Error(err)
		}
		return
	}
	referer := r.Header.Get("referer")
	link.Clicks++
	if referer == "" {
		link.ClicksNone++
	} else if strings.Contains(referer, "facebook.com") {
		link.ClicksFacebook++
	} else if strings.Contains(referer, "instagram.com") {
		link.ClicksInstagram++
	} else {
		link.ClicksOther++
	}
	err = connection.Collection("links").Save(link)
	if err != nil {
		logrus.Error(err)
	}
	requestTime := time.Since(requestTimer)
	logrus.Info(fmt.Sprintf("[%v] Redirecting %s to %s", requestTime, link.Name, link.Url))
	http.Redirect(w, r, link.Url, 302)
}

func newShortUrl(w http.ResponseWriter, r *http.Request) {
	requestTimer := time.Now()
	if err := r.ParseForm(); err != nil {
		returnError500(err, w)
		return
	}

	url := addHttp(r.FormValue("url"))
	emoji := r.FormValue("emoji")
	password := r.FormValue("password")
	var response string
	if r.FormValue("accept") != "" {
		response = r.FormValue("accept")
	} else {
		response = strings.Split(r.Header.Get("Accept"), ",")[0]
	}
	var name string
	link := &Link{}
	if password != "" && password == viper.GetString("admin-password") && r.FormValue("name") != "" {
		name = r.FormValue("name")
		err := connection.Collection("links").FindOne(bson.M{"name": name}, link)
		if err != nil {
			link = &Link{
				Url:url,
				Name:name,
			}
		}
		link.Url = url
	} else {
		var err error
		name, err = getUniqueRandomString(6, emoji == "1")
		if err != nil {
			returnError500(err, w)
			return
		}
		link = &Link{
			Url:url,
			Name:name,
		}
	}
	link.Password = randomString(16, false)
	err := connection.Collection("links").Save(link)
	if err != nil {
		returnError500(err, w)
		return
	}
	switch response {
	case "application/json":
		b, err := json.Marshal(JsonResponse{
			Shorturl:  fmt.Sprintf("%s/%s", viper.GetString("base-url"), name),
			Url:       url,
			Manageurl: fmt.Sprintf("%s/%s/%s", viper.GetString("base-url"), name, link.Password),
		})
		if err != nil {
			returnError500(err, w)
			return
		}
		w.Write(b)
		break
	case "text/plain":
		_, _ = fmt.Fprintf(w, "%s/%s", viper.GetString("base-url"), name)
		break
	default:
		err = templates.ExecuteTemplate(w, "new.html", map[string]interface{}{
			"BaseUrl": viper.GetString("base-url"),
			"Link": link,
		})
		if err != nil {
			logrus.Error(err)
		}
		break
	}
	requestTime := time.Since(requestTimer)
	logrus.Info(fmt.Sprintf("[%v] New Shorturl: %s redirects to %s (%s)", requestTime, name, url, r.Header.Get("Accept")))
}

func monitoringHandler(w http.ResponseWriter, r *http.Request) {
	link := &Link{}
	err := connection.Collection("links").FindOne(bson.M{}, link)
	if err != nil {
		returnError500(err, w)
		return
	}
	w.WriteHeader(200)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	returnError404(w)
	logrus.Info(fmt.Sprintf("URL not found: %s %s", r.URL, r.Method))
}
