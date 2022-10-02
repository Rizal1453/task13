package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"main.go/connection"
)
func main (){
	route := mux.NewRouter()
	connection.DatabaseConnect()

	route.HandleFunc("/", index).Methods("GET")
	route.HandleFunc("/add", Add).Methods("POST")
	route.HandleFunc("/delete/{id}", delete).Methods("GET")
	
	route.HandleFunc("/registrasi", registrasi).Methods("GET")
	route.HandleFunc("/submitregistrasi", submitregistrasi).Methods("POST")

	route.HandleFunc("/login", login).Methods("GET")
	route.HandleFunc("/submitlogin", submitlogin).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	

	fmt.Println("server started at localhost:8000")
	http.ListenAndServe("localhost:9000", route)
}
type SessionData struct{
	IsLogin bool
	UserName string
}

var Data = SessionData{}

type User struct{
	ID  int
	Name string
	Email string
	Password string
}
type Project struct{
	Title string
	Content string
	ID int
}
func index(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html; charset=utf8")
	var tmpl, err = template.ParseFiles("index.html")

	if err != nil{
		w.Write([]byte("web tidak tersedia" + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	data,_ :=connection.Conn.Query(context.Background(),"SELECT tb_todo.id, tb_todo.title, Content FROM tb_todo ORDER BY id DESC")
	var result[] Project
	for data.Next(){
		var each = Project{}
		err:= data.Scan(&each.ID,&each.Title,&each.Content)
		if err != nil{
			fmt.Println(err.Error())
			return
		}
		result = append(result, each)
	}
	resData :=map[string]interface{}{
		"Blogs":result,
		"Data":Data,
	}
	fmt.Println()
		tmpl.Execute(w,resData)

}
func Add(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")

	fmt.Println(title)
	fmt.Println(content)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_todo (title,content) VALUES ($1, $2)", title, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : "+ err.Error()))
		return}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func registrasi(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html; charset=utf8")
	var tmpl, err = template.ParseFiles("registration.html")

	if err != nil{
		w.Write([]byte("web tidak tersedia" + err.Error()))
		return
	}
	tmpl.Execute(w,nil)
}
func submitregistrasi(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")
	fmt.Println(password)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	fmt.Println(passwordHash)
	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func login(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html; charset=utf8")
	var tmpl, err = template.ParseFiles("login.html")

	if err != nil{
		w.Write([]byte("message : "))
	}
	tmpl.Execute(w,nil)
}
func submitlogin(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")


	user := User{}

	// mengambil data email, dan melakukan pengecekan email
	err = connection.Conn.QueryRow(context.Background(),
		"SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		fmt.Println("Email belum terdaftar")
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	
		return
	}

	// melakukan pengecekan password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Println("Password salah")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : Password salah " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 10800 // 3 JAM

	session.AddFlash("succesfull","message")
	session.Save(r,w)
	

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
func delete(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_todo WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
func logout(w http.ResponseWriter, r *http.Request){
	fmt.Println("logout")
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session,_ := store.Get(r,"SESSION_KEY")
	session.Options.MaxAge= -1
	session.Save(r,w)

	http.Redirect(w,r,"/login",http.StatusSeeOther)
}