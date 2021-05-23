package heroku

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/OloloevReal/go-simple-log"
)

type AppHeroku struct {
	ServiceURL   string
	Port         string
	Ctx          context.Context
	responseText string
}

func (app *AppHeroku) RunKeepAlive() {
	if len(app.responseText) == 0 {
		app.responseText = "Hi there!"
	}
	app.runHerokuAlive(app.Ctx)
	app.runHerokuHandler(app.Ctx)
}

func (app *AppHeroku) SetText(text string) {
	app.responseText = text
}

func (app *AppHeroku) runHerokuAlive(ctx context.Context) {
	if app.ServiceURL == "" {
		return
	}
	var t *time.Ticker
	client := http.Client{}
	log.Printf("Starting Heroku alive func, %s", app.ServiceURL)
	go func() {
		defer log.Printf("Closed Heroku alive func, %s", app.ServiceURL)
		for {
			t = time.NewTicker(10 * time.Minute)
			select {
			case <-t.C:
				{
					client.Get(app.ServiceURL)
				}
			case <-ctx.Done():
				{
					return
				}
			}

		}
	}()
}

func (app *AppHeroku) runHerokuHandler(ctx context.Context) {
	log.Println("Starting Heroku http handler")
	fnHandler := func(resp http.ResponseWriter, _ *http.Request) {
		resp.Write([]byte(app.responseText))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", fnHandler)
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", app.Port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	//TODO: GO1.13 errors.Is
	go func() {
		defer log.Println("Closed Heroku http handler")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Println("[ERROR] failed close http server, %v", err)
		}
	}()
}
