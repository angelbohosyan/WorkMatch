package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func getCompleteMatches(user string) map[string]Task {
	query := fmt.Sprintf("Select * from sys.completematches where completematches.username = '%s'", user)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task, 100)
	id := 0
	for rows.Next() {
		var (
			description string
			language    string
			username    string
			link        string
		)
		if err = rows.Scan(&description, &language, &username, &link); err != nil {
			panic(err)
		}
		tasks[strconv.Itoa(id)] = Task{
			Description: description,
			Language:    language,
			Username:    username,
			IsLink:      false,
		}
		id++
	}
	return tasks
}
func getCompleteMatchesForMe(user string) map[string]Task {
	query := fmt.Sprintf("Select * from sys.completematches where completematches.username = '%s'", user)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task, 100)
	id := 0
	for rows.Next() {
		var (
			description    string
			language       string
			username       string
			completeTaskId string
		)
		if err = rows.Scan(&description, &language, &username, &completeTaskId); err != nil {
			panic(err)
		}
		tasks[strconv.Itoa(id)] = Task{
			Description: description,
			Language:    language,
			Username:    username,
			Link:        completeTaskId,
			IsLink:      true,
		}
		id++
	}
	return tasks
}

func getTaskIDWithUser(user string) string {
	query := fmt.Sprintf("Select TaskID from sys.tasks where tasks.username = '%s'", user)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var username string

	if err = value.Scan(&username); err != nil {
		panic(err)
	}
	return username
}

func checkMatchIfDone(username string, id string) {
	query := fmt.Sprintf("Select done1 from sys.matches where matches.task1 = '%s'", id)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if value.Next() {
		var password bool
		if err = value.Scan(&password); err != nil {
			panic(err)
		}
		if password {
			query2 := fmt.Sprintf("Select task2, username2 from sys.matches where matches.task1 = '%s'", id)
			value2, err2 := DB.Query(query2)
			if err2 != nil {
				log.Fatal(err2)
			}
			defer value2.Close()
			value2.Next()
			var task2 string
			var username2 string
			if err = value2.Scan(&task2, &username2); err != nil {
				panic(err)
			}
			completeMatch(id, username, task2)
			completeMatch(task2, username2, id)
			removeMatch(id)
		} else {
			query = fmt.Sprintf("Update sys.matches set done2=%s where matches.task1='%s'", "true", id)
			fmt.Println(query)
			_, err = DB.Query(query)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	query2 := fmt.Sprintf("Select done2 from sys.matches where matches.task2 = '%s'", id)
	value2, err2 := DB.Query(query2)
	if err2 != nil {
		log.Fatal(err2)
	}
	defer value2.Close()
	if value2.Next() {
		var password2 bool
		if err = value2.Scan(&password2); err != nil {
			panic(err)
		}
		if password2 {
			query = fmt.Sprintf("Select task1,username1 from sys.matches where matches.task2 = '%s'", id)
			value3, err3 := DB.Query(query)
			if err3 != nil {
				log.Fatal(err3)
			}
			defer value3.Close()
			value3.Next()
			var task2 string
			var username2 string
			if err = value3.Scan(&task2, &username2); err != nil {
				panic(err)
			}
			completeMatch(id, username, task2)
			completeMatch(task2, username2, id)
			removeMatch(id)
		} else {
			query = fmt.Sprintf("Update sys.matches set done1=%s where matches.task2='%s'", "true", id)
			_, err = DB.Query(query)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func checkIfLikeExists(user string, id string) bool {
	query := fmt.Sprintf("Select if( exists(Select * from sys.likedtasks where likedtasks.likeusername = '%s' AND likedtasks.TaskID = '%s'),true,false)", user, id)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var password bool
	if err := value.Scan(&password); err != nil {
		panic(err)
	}
	if password {
		return true
	} else {
		return false
	}
}

func returnMatches(user string) map[string]Task {
	query := fmt.Sprintf("Select task2 from sys.matches where matches.username2 = '%s' union Select task1 from sys.matches where matches.username1 = '%s'", user, user)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task, 100)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			panic(err)
		}
		query2 := fmt.Sprintf("Select * from sys.tasksmatches where tasksmatches.TaskID = '%s'", id)
		rows2, err2 := DB.Query(query2)
		if err2 != nil {
			log.Fatal(err2)
		}
		for rows2.Next() {
			var (
				id2         string
				description string
				language    string
				date        string
				username    string
			)
			if err = rows2.Scan(&id2, &description, &language, &date, &username); err != nil {
				panic(err)
			}
			tasks[id] = Task{
				ID:          id2,
				Description: description,
				Language:    language,
				Time:        date,
				Username:    username,
				Email:       getEmailByUsername(username),
			}
		}
		rows2.Close()
	}
	return tasks
}

func getEmailByUsername(username string) string {
	query := fmt.Sprintf("Select email from sys.user where user.username = '%s'", username)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var email string

	if err = value.Scan(&email); err != nil {
		panic(err)
	}
	return email
}

func returnAdminMatches() map[int]Match {
	query := "Select * from sys.matches"
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[int]Match, 100)
	i := -1
	for rows.Next() {
		var (
			id1       string
			username  string
			id2       string
			username2 string
			bool1     bool
			bool2     bool
		)
		if err = rows.Scan(&id1, &username, &id2, &username2, &bool1, &bool2); err != nil {
			panic(err)
		}
		id1, description1, language1, date1, username1 := getMatchTaskInformationByID(id1)
		id2, description2, language2, date2, username2 := getMatchTaskInformationByID(id2)
		i++
		tasks[i] = Match{
			ID1:          id1,
			Language1:    language1,
			Description1: description1,
			Time1:        date1,
			Username1:    username1,
			ID2:          id2,
			Language2:    language2,
			Description2: description2,
			Time2:        date2,
			Username2:    username2,
		}
	}
	return tasks
}

func removeTask(id string) {
	query := fmt.Sprintf("Delete from sys.tasks where tasks.TaskID = '%s'", id)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	removeLikeTask(id)
}

func removeMatch(id string) {
	query := fmt.Sprintf("Delete from sys.matches where matches.task1 = '%s' OR matches.task2 = '%s'", id, id)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func completeMatch(id string, user string, completedId string) {
	description, language := showTaskInformationByID(id)
	query := fmt.Sprintf("Insert into sys.completematches (TaskDescription,TaskLanguage,username,id) values ('%s' , '%s' , '%s','%s')", description, language, user, completedId)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertUser(email string, name string, password string) {
	query := fmt.Sprintf("Insert into sys.user (username,password,email) values ('%s' , '%s','%s')", name, password, email)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func checkForUser(name string, sentPassword string) bool {
	query := fmt.Sprintf("Select password from sys.user where user.username = '%s'", name)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if !value.Next() {
		return false
	}
	var password string
	if err := value.Scan(&password); err != nil {
		panic(err)
	}
	if password == sentPassword {
		return true
	} else {
		return false
	}
}

func checkForAdminUser(name string) bool {
	query := fmt.Sprintf("Select password from sys.user where user.username = '%s'", name)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if !value.Next() {
		return false
	} else {
		return true
	}
}

func checkForRegisteredUser(name string) bool {
	query := fmt.Sprintf("Select if( exists(Select username from sys.user where user.username = '%s'),true,false)", name)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var password bool
	if err := value.Scan(&password); err != nil {
		panic(err)
	}
	if password {
		log.Printf("There is already a user with the name %s\n", name)
		return true
	} else {
		return false
	}
}

func insertTask(description string, language string, date string, user string) {
	language = strings.ToLower(language)
	query := fmt.Sprintf("Insert into sys.tasks (TaskDescription,TaskLanguage,deadline,username) values ('%s' , '%s' , '%s','%s')", description, language, date, user)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertTaskMatches(id string) {
	id, description, language, date, username := getTaskInformationByID(id)
	query := fmt.Sprintf("Insert into sys.tasksmatches (TaskID,TaskDescription,TaskLanguage,deadline,username) values ('%s','%s' , '%s' , '%s','%s')", id, description, language, date, username)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func getTaskInformationByID(id2 string) (string, string, string, string, string) {
	query := fmt.Sprintf("Select * from sys.tasks where TaskID='%s'", id2)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	value.Next()
	var (
		id          string
		description string
		language    string
		date        string
		username    string
	)
	if err = value.Scan(&id, &description, &language, &date, &username); err != nil {
		panic(err)
	}
	return id, description, language, date, username
}

func getMatchTaskInformationByID(id2 string) (string, string, string, string, string) {
	query := fmt.Sprintf("Select * from sys.tasksmatches where TaskID='%s'", id2)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	value.Next()
	var (
		id          string
		description string
		language    string
		date        string
		username    string
	)
	if err = value.Scan(&id, &description, &language, &date, &username); err != nil {
		panic(err)
	}
	return id, description, language, date, username
}

func showTaskInformation(user string) map[string]Task2 {
	rows, err := DB.Query("SELECT * FROM sys.tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task2, 10)
	for rows.Next() {
		var (
			id          string
			description string
			language    string
			date        string
			username    string
		)
		if err := rows.Scan(&id, &description, &language, &date, &username); err != nil {
			panic(err)
		}
		tasks[id] = Task2{
			ID:          id,
			Description: description,
			Language:    language,
			Time:        date,
			Username:    username,
			Email:       getEmailByUsername(username),
			IsLiked:     checkIfLikeExists(user, id),
		}
	}
	return tasks
}

func showLanguageTaskInformation(user string, language string) map[string]Task2 {
	query := fmt.Sprintf("Select * from sys.tasks where TaskLanguage='%s'", language)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task2, 10)
	for rows.Next() {
		var (
			id          string
			description string
			language2   string
			date        string
			username    string
		)
		if err = rows.Scan(&id, &description, &language2, &date, &username); err != nil {
			panic(err)
		}
		tasks[id] = Task2{
			ID:          id,
			Description: description,
			Language:    language,
			Time:        date,
			Username:    username,
			IsLiked:     checkIfLikeExists(user, id),
		}
	}
	return tasks
}

func showTaskInformationByID(id string) (string, string) {
	query := fmt.Sprintf("SELECT * FROM sys.tasksmatches where TaskID='%s'", id)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	rows.Next()
	var (
		id2         string
		description string
		language    string
		date        string
		username    string
	)
	if err = rows.Scan(&id2, &description, &language, &date, &username); err != nil {
		panic(err)
	}

	query = fmt.Sprintf("Delete from sys.tasksmatches where tasksmatches.TaskID = '%s'", id)
	_, err = DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	removeLikeTask(id)

	return description, language
}

func showAdminTaskInformation() map[string]Task {
	rows, err := DB.Query("SELECT * FROM sys.tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task, 10)
	for rows.Next() {
		var (
			id          string
			description string
			language    string
			date        string
			username    string
		)
		if err := rows.Scan(&id, &description, &language, &date, &username); err != nil {
			panic(err)
		}
		tasks[id] = Task{
			ID:          id,
			Description: description,
			Language:    language,
			Time:        date,
			Username:    username,
		}
	}
	return tasks
}

func showYourTaskInformation(user string) map[string]Task {
	query := fmt.Sprintf("Select * from sys.tasks where tasks.username='%s'", user)
	rows, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tasks := make(map[string]Task, 10)
	for rows.Next() {
		var (
			id          string
			description string
			language    string
			date        string
			username    string
		)
		if err = rows.Scan(&id, &description, &language, &date, &username); err != nil {
			panic(err)
		}
		tasks[id] = Task{
			ID:          id,
			Description: description,
			Language:    language,
			Time:        date,
			Username:    username,
			Email:       getEmailByUsername(username),
		}
	}
	return tasks
}

func likeTask(likeUsername string, id string) {
	username := getUserWithTaskID(id)
	if username == likeUsername {
		return
	}
	if !checkIfLikeExists(likeUsername, id) {
		query := fmt.Sprintf("Insert into sys.likedtasks (TaskID,username,likeusername) values ('%s' , '%s', '%s')", id, username, likeUsername)
		_, err := DB.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		checkForMatch(username, likeUsername)
	} else {
		fmt.Println("Like already exists")
	}
}

func getTasksID(username1, username2 string) (string, string) {
	query := fmt.Sprintf("Select TaskID from sys.likedtasks where likedtasks.likeusername = '%s' AND likedtasks.username='%s'", username1, username2)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	query2 := fmt.Sprintf("Select TaskID from sys.likedtasks where likedtasks.likeusername = '%s' AND likedtasks.username='%s'", username2, username1)
	value2, err2 := DB.Query(query2)
	if err2 != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var id1 string
	if err = value.Scan(&id1); err != nil {
		panic(err)
	}
	defer value2.Close()
	value2.Next()
	var id2 string
	if err = value2.Scan(&id2); err != nil {
		panic(err)
	}
	return id1, id2
}

func checkForMatch(username string, likeUsername string) {
	matchQuery := fmt.Sprintf("Select * from sys.likedtasks where likedtasks.likeusername = '%s' AND likedtasks.username='%s'", username, likeUsername)
	value, err := DB.Query(matchQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if value.Next() {
		fmt.Println("There is a match")
		id1, id2 := getTasksID(username, likeUsername)
		insertMatch(id1, username, id2, likeUsername)
	}
}

func insertMatch(id1 string, username1 string, id2 string, username2 string) {
	query := fmt.Sprintf("Insert into sys.matches (task1,username1,task2,username2) values ('%s' , '%s' , '%s','%s')", id1, username1, id2, username2)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	insertTaskMatches(id1)
	insertTaskMatches(id2)
	removeTask(id1)
	removeTask(id2)
}

func banUser(username string) {
	query := fmt.Sprintf("Insert into sys.banemails (email) values ('%s')", getEmailByUsername(username))
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	query = fmt.Sprintf("Delete from sys.user where user.username = '%s'", username)
	_, err = DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	query = fmt.Sprintf("Delete from sys.tasks where tasks.username = '%s'", username)
	_, err = DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	query = fmt.Sprintf("Select task1,task2 from sys.matches  where matches.username1 = '%s' OR matches.username2 = '%s'", username, username)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	for value.Next() {
		var (
			id1 string
			id2 string
		)
		if err = value.Scan(&id1, &id2); err != nil {
			panic(err)
		}
		query = fmt.Sprintf("Delete from sys.tasksmatches where tasksmatches.TaskID = '%s' OR tasksmatches.TaskID = '%s'", id1, id2)
		_, err = DB.Query(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	query = fmt.Sprintf("Delete from sys.matches where matches.username1 = '%s' OR matches.username2 = '%s'", username, username)
	_, err = DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	query = fmt.Sprintf("Delete from sys.likedtasks where likedtasks.username = '%s' OR likedtasks.likeusername = '%s'", username, username)
	_, err = DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func emailIsBanned(email string) bool {
	query := fmt.Sprintf("Select * from sys.banemails where banemails.email = '%s'", email)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if value.Next() {
		return true
	} else {
		return false
	}
}

func getUserWithTaskID(id string) string {
	query := fmt.Sprintf("Select username from sys.tasks where tasks.TaskID = '%s'", id)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var username string

	if err := value.Scan(&username); err != nil {
		panic(err)
	}
	return username
}

func checkForRegisteredAdmin(name string) bool {
	query := fmt.Sprintf("Select if( exists(Select username from sys.admins where admins.username = '%s'),true,false)", name)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	value.Next()
	var password bool
	if err := value.Scan(&password); err != nil {
		panic(err)
	}
	if password {
		log.Printf("There is already a admin with the name %s\n", name)
		return true
	} else {
		return false
	}
}

func checkForAdmin(name string, sentPassword string) bool {
	query := fmt.Sprintf("Select password from sys.admins where admins.username = '%s'", name)
	value, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer value.Close()
	if !value.Next() {
		return false
	}
	var password string
	if err := value.Scan(&password); err != nil {
		panic(err)
	}
	if password == sentPassword {
		return true
	} else {
		return false
	}
}

func removeLikeTask(id string) {
	query := fmt.Sprintf("Delete from sys.likedtasks where likedtasks.TaskID = '%s'", id)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}

func removeLikeTaskFromUser(id string, user string) {
	query := fmt.Sprintf("Delete from sys.likedtasks where likedtasks.TaskID = '%s' AND likedtasks.likeusername= '%s'", id, user)
	_, err := DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
}
