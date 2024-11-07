package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/febriW/be-to-do/card"
	"github.com/febriW/be-to-do/user"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s", r.Method, r.URL.String()+" Not Exist")
}

func main() {
	db := initDB()
	userService := user.NewService(db)
	cardService := card.NewService(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", NotImplemented)
	mux.HandleFunc("POST /user/register", userService.HandleRegister())
	mux.HandleFunc("POST /auth/login", userService.HandleLogin())

	mux.HandleFunc("GET /card", user.TokenMiddleware(cardService.HandleGetAllCards()))
	mux.HandleFunc("POST /card", user.TokenMiddleware(cardService.HandleCreateCard()))
	mux.HandleFunc("PUT /card", user.TokenMiddleware(cardService.HandleUpdateCard()))
	mux.HandleFunc("DELETE /card", user.TokenMiddleware(cardService.HandleDeleteCard()))

	srv := &http.Server{
		Handler:      mux,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Println("Server is running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("got interruption signal")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown returned err: %v", err)
	}

}

func initDB() *sql.DB {
	db, err := sql.Open("mysql", "root:abc123@tcp(db:3306)/appdb?parseTime=true&loc=Asia%2FJakarta")
	if err != nil {
		panic(err)
	}

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}
