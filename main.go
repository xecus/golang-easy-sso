package main

import (
	"fmt"
	"github.com/StephanDollberg/go-json-rest-middleware-jwt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

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
	hostname := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DATABASE")
	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, hostname, port, database)
	i.DB, err = gorm.Open("postgres", uri)
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

func SetCors(api *rest.Api, allow_origin string) {
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == allow_origin
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
}

func GenerateJwtMiddleware(i *Impl) *jwt.JWTMiddleware {

	return &jwt.JWTMiddleware{
		Key:        []byte("secret key"),
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(username string, password string) bool {
			var user User
			if err := i.DB.Where(&User{Username: username, Password: password}).First(&user).Error; err != nil {
				log.Println("[Auth] unknown user")
				return false
			}
			if !user.Enabled {
				log.Println("[Auth] Unenabled user")
				return false
			}
			user.LastUseAt = time.Now()
			if err := i.DB.Save(&user).Error; err != nil {
				log.Println("Error: could not update user model")
				return false
			}
			log.Println("[Auth] OK")
			return true
		}}
}

func main() {

	// Prepare
	log.Println("[Main] loading .env file")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[Main] Error loading .env file")
	}

	log.Println("[Main] initilizing database")
	i := Impl{}
	i.InitDB()
	i.InitSchema()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	// CORS
	if os.Getenv("ENABLE_CORS") == "true" {
		allow_origin := os.Getenv("ALLOW_ORIGIN")
		log.Println("[Main] allow host " + allow_origin)
		SetCors(api, allow_origin)
	}

	// JWT Generator
	jwt_middleware := GenerateJwtMiddleware(&i)
	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			authRequest := request.URL.Path == "/auth"
			userControlRequest := strings.HasPrefix(request.URL.Path, "/users")
			return !(authRequest || userControlRequest)
		},
		IfTrue: jwt_middleware,
	})

	// Serve
	router, _ := rest.MakeRouter(
		rest.Get("/users", i.GetAllUsers),
		rest.Post("/users", i.PostUser),
		rest.Get("/users/:id", i.GetUser),
		rest.Put("/users/:id", i.UpdateUser),
		rest.Delete("/users/:id", i.DeleteUser),
		rest.Post("/auth", jwt_middleware.LoginHandler),
		rest.Get("/refresh", jwt_middleware.RefreshHandler),
	)
	api.SetApp(router)
	http.Handle("/api/v1/", http.StripPrefix("/api/v1", api.MakeHandler()))
	port := os.Getenv("SERVICE_PORT")
	log.Println("[Main] Starting Server Port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
