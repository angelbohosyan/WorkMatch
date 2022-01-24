package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key          = []byte("super-secret-key")
	store        = sessions.NewCookieStore(key)
	DB           *sql.DB
	tmplAllTasks *template.Template
	workDir, _   = os.Getwd()
)

func init() {
	var tmplBase = template.New("base").Funcs(
		template.FuncMap{
			"urlSafe": func(url url.URL) template.HTML {
				return template.HTML(url.String())
			},
		})

	var err error
	tmplAllTasks, err = tmplBase.ParseFiles(
		path.Join(workDir, "templates", "matches.html"),
		path.Join(workDir, "templates", "tasks.html"),
		path.Join(workDir, "templates", "login.html"),
		path.Join(workDir, "templates", "maketask.html"),
		path.Join(workDir, "templates", "matchesadmin.html"),
		path.Join(workDir, "templates", "tasksadmin.html"),
		path.Join(workDir, "templates", "yourtasks.html"),
		path.Join(workDir, "templates", "register.html"),
		path.Join(workDir, "templates", "completematches.html"),
		path.Join(workDir, "templates", "useradmin.html"),
	)
	if err != nil {
		log.Println(err)
	}
	for _, t := range tmplAllTasks.Templates() {
		_, err = t.ParseFiles(
			path.Join(workDir, "templates", "head.html"),
			path.Join(workDir, "templates", "nav.html"),
			path.Join(workDir, "templates", "nav2.html"),
		)
		if err != nil {
			log.Println(err)
		}
	}
	cfg := mysql.Config{
		User:   "root",
		Passwd: "2805",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "sys",
	}

	// Get a database handle.
	DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
}

func showMatches(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := session.Values["user"].(string)
	taskID := req.FormValue("id")
	if taskID != "" {
		checkMatchIfDone(user, taskID)
	}
	matches := returnMatches(user)
	err := tmplAllTasks.ExecuteTemplate(w, "matches.html", matches)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func showCompleteMatches(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := req.FormValue("username")
	if user == session.Values["user"].(string) {
		matches := getCompleteMatchesForMe(user)
		err := tmplAllTasks.ExecuteTemplate(w, "completematches.html", matches)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	} else {
		matches := getCompleteMatches(user)
		err := tmplAllTasks.ExecuteTemplate(w, "completematches.html", matches)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	}
}

func showAdminCompleteMatches(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["adminAuthenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := req.FormValue("username")
	ban := req.FormValue("banusername")
	if ban != "" {
		if checkForAdminUser(user) {
			banUser(user)
		}
		err := tmplAllTasks.ExecuteTemplate(w, "useradmin.html", nil)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	} else {
		matches := getCompleteMatches(user)
		err := tmplAllTasks.ExecuteTemplate(w, "useradmin.html", matches)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	}

}

func showAdminMatches(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["adminAuthenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	taskID := req.FormValue("id1")
	if taskID != "" {
		if req.FormValue("removeid") != "" {
			removeMatch(taskID)
			showTaskInformationByID(req.FormValue("id1"))
			showTaskInformationByID(req.FormValue("id2"))
		} else {
			if req.FormValue("completeid") != "" {
				checkMatchIfDone(req.FormValue("name1"), req.FormValue("id2"))
				checkMatchIfDone(req.FormValue("name2"), req.FormValue("id1"))
			}
		}
	}
	matches := returnAdminMatches()

	err := tmplAllTasks.ExecuteTemplate(w, "matchesadmin.html", matches)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func showTasks(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := session.Values["user"].(string)
	taskID := req.FormValue("id")
	if taskID != "" {
		likeTask(user, taskID)
	}
	taskID = req.FormValue("dislikeid")
	if taskID != "" {
		removeLikeTaskFromUser(taskID, user)
	}
	language := req.FormValue("language")
	if language != "" {
		tasks := showLanguageTaskInformation(user, language)
		err := tmplAllTasks.ExecuteTemplate(w, "tasks.html", tasks)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	} else {
		tasks := showTaskInformation(user)
		err := tmplAllTasks.ExecuteTemplate(w, "tasks.html", tasks)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
		}
	}
}

func showAdminTasks(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["adminAuthenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	taskID := req.FormValue("id")
	if taskID != "" {
		removeTask(taskID)
	}
	tasks := showAdminTaskInformation()
	err := tmplAllTasks.ExecuteTemplate(w, "tasksadmin.html", tasks)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func makeTask(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := session.Values["user"].(string)
	description := req.FormValue("description")
	language := req.FormValue("language")
	date := req.FormValue("date")
	if description != "" && language != "" && date != "" {
		insertTask(description, language, date, user)
	}
	err := tmplAllTasks.ExecuteTemplate(w, "maketask.html", nil)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func showYourTasks(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user := session.Values["user"].(string)
	taskID := req.FormValue("id")
	if taskID != "" {
		removeTask(taskID)
	}
	tasks := showYourTaskInformation(user)
	err := tmplAllTasks.ExecuteTemplate(w, "yourtasks.html", tasks)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func registerUser(w http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")
	username := req.FormValue("username")
	password := req.FormValue("password")
	var description string
	if email != "" && username != "" && password != "" {
		if !checkForRegisteredUser(username) && !checkForRegisteredAdmin(username) {
			if !emailIsBanned(email) {
				if _, err := mail.ParseAddress(email); err == nil {
					insertUser(email, username, password)
				} else {
					description = "not valid email"
				}
			} else {
				description = "this email is banned"
			}
		} else {
			description = "There is already a user with that name"
		}
	}
	err := tmplAllTasks.ExecuteTemplate(w, "register.html", description)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
	}
}

func login(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	password := req.FormValue("password")
	description := ""
	isLogout := req.FormValue("logout")
	session, _ := store.Get(req, "sessionID")
	if isLogout == "true" {
		session.Values["authenticated"] = false
		session.Values["adminAuthenticated"] = false
		session.Values["user"] = ""
		err := session.Save(req, w)
		if err != nil {
			return
		}
	}
	// Check if user is authenticated

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		showMatches(w, req)
	} else {
		if auth, ok = session.Values["adminAuthenticated"].(bool); ok && auth {
			showAdminMatches(w, req)
		} else {
			if username != "" && password != "" {
				if checkForUser(username, password) {
					session.Values["authenticated"] = true
					session.Values["user"] = username
					err := session.Save(req, w)
					if err != nil {
						return
					}
					showMatches(w, req)
					return
				} else {
					if checkForAdmin(username, password) {
						session.Values["adminAuthenticated"] = true
						session.Values["user"] = username
						err := session.Save(req, w)
						if err != nil {
							return
						}
						showAdminMatches(w, req)
						return
					} else {
						description = "There is no such user"
					}
				}
			}
			err := tmplAllTasks.ExecuteTemplate(w, "login.html", description)
			if err != nil {
				log.Printf("Error executing template: %v\n", err)
			}
		}
	}
}

func uploadFile(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "sessionID")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	req.ParseMultipartForm(10 << 20)
	file, _, err := req.FormFile("myFile")
	id := req.FormValue("uploadId")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	name := fmt.Sprintf("%s/files/%s.zip", workDir, id)
	tempFile, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		return
	}
	fmt.Println("Successfully Uploaded File")
	showMatches(w, req)

}

func main() {
	var addr = flag.String("addr", ":8080", "http service address")
	http.HandleFunc("/register", registerUser)
	http.HandleFunc("/completematches", showCompleteMatches)
	http.HandleFunc("/maketask", makeTask)
	http.HandleFunc("/matches", showMatches)
	http.HandleFunc("/tasks", showTasks)
	http.HandleFunc("/tasksadmin", showAdminTasks)
	http.HandleFunc("/matchesadmin", showAdminMatches)
	http.HandleFunc("/yourtasks", showYourTasks)
	http.HandleFunc("/useradmin", showAdminCompleteMatches)
	http.HandleFunc("/uploadfile", uploadFile)
	http.HandleFunc("/", login)
	fs := http.FileServer(http.Dir(path.Join(workDir, "static")))
	fs2 := http.FileServer(http.Dir(path.Join(workDir, "files")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/files/", http.StripPrefix("/files/", fs2))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
