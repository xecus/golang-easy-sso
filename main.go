package main

import (
	"fmt"
	"github.com/StephanDollberg/go-json-rest-middleware-jwt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"time"
)

func handle_auth(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})
}

type User struct {
	gorm.Model
	Username  string    `gorm:"size:128" json:"username"`
	Password  string    `gorm:"size:128" json:"password"`
	Enabled   bool      `json:"enabled"`
	LastUseAt time.Time `json:"lastUseAt"`
}

type Impl struct {
	DB *gorm.DB
}

func (i *Impl) InitDB() {
	var err error
	i.DB, err = gorm.Open("postgres", "host=localhost user=postgres dbname=taguro sslmode=disable password=postgres")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
	//defer i.DB.Close()
	//i.DB.LogMode(true)
}

func (i *Impl) InitSchema() {
	i.DB.AutoMigrate(&User{})
}

func (i *Impl) GetAllUsers(w rest.ResponseWriter, r *rest.Request) {
	users := []User{}
	i.DB.Find(&users)
	w.WriteJson(&users)
}

func (i *Impl) GetUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := User{}
	if i.DB.First(&user, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&user)
}

func (i *Impl) PostUser(w rest.ResponseWriter, r *rest.Request) {
	user := User{}
	if err := r.DecodeJsonPayload(&user); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := i.DB.Save(&user).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&user)
}

func (i *Impl) UpdateUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := User{}
	if i.DB.First(&user, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	updated := User{}
	if err := r.DecodeJsonPayload(&updated); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Username = updated.Username
	user.Password = updated.Password
	user.Enabled = updated.Enabled
	if err := i.DB.Save(&user).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&user)
}

func (i *Impl) DeleteUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := User{}
	if i.DB.First(&user, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	if err := i.DB.Delete(&user).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {

	i := Impl{}
	i.InitDB()
	i.InitSchema()

	jwt_middleware := &jwt.JWTMiddleware{
		Key:        []byte("secret key"),
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(username string, password string) bool {
			var user User
			if err := i.DB.Where(&User{Username: username, Password: password}).First(&user).Error; err != nil {
				fmt.Println("Error: Not found user record")
				return false
			}
			if user.Enabled {
				fmt.Println("Enabled = true")
				user.LastUseAt = time.Now()
				if err := i.DB.Save(&user).Error; err != nil {
					fmt.Println("Error: could not update user model")
					return false
				}
				return true
			} else {
				fmt.Println("Enabled = false")
				return false
			}
			return false
		}}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://25.37.37.128:38080"
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			// return request.URL.Path != "/login"
			return false
		},
		IfTrue: jwt_middleware,
	})
	router, _ := rest.MakeRouter(

		rest.Get("/users", i.GetAllUsers),
		rest.Post("/users", i.PostUser),
		rest.Get("/users/:id", i.GetUser),
		rest.Put("/users/:id", i.UpdateUser),
		rest.Delete("/users/:id", i.DeleteUser),

		rest.Post("/login", jwt_middleware.LoginHandler),
		rest.Get("/auth_test", handle_auth),
		rest.Get("/refresh_token", jwt_middleware.RefreshHandler),
	)
	api.SetApp(router)

	http.Handle("/api/v1/", http.StripPrefix("/api/v1", api.MakeHandler()))

	fmt.Println("Starting Server Port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
