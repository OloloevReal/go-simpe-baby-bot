package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/OloloevReal/go-simple-baby-bot/goenvs"
	"github.com/OloloevReal/go-simple-baby-bot/heroku"
	"github.com/OloloevReal/go-simple-baby-bot/store"
	log "github.com/OloloevReal/go-simple-log"
)

const version = "0.0.1"

type AppEnvs struct {
	Port          string `env:"PORT"`
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	TelegramProxy string `env:"TELEGRAM_PROXY"`
	MongoURL      string `env:"MONGO_URL"`
	MongoDB       string `env:"MONGO_DB"`
	StoreType     string `env:"STORE_TYPE"`
	ServiceURL    string `env:"SERVICE_URL"`
	Debug         bool   `env:"DEBUG"`
}

type App struct {
	envs  *AppEnvs
	store store.Store
}

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected")
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}

func main() {
	log.Printf("Started go-baby-bot version %s\r\n", version)
	defer log.Println("Finished!")

	envs := new(AppEnvs)
	if err := goenvs.Parse(envs); err != nil {
		log.Fatalf("[FATAL] parsing envs failed, %s", err)
	}

	if envs.Debug {
		log.SetOptions(log.SetDebug)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[INFO] interrupt signal")
		cancel()
	}()

	_ = ctx

	herroku := heroku.AppHeroku{
		ServiceURL: envs.ServiceURL,
		Port:       envs.Port,
		Ctx:        ctx,
	}

	if len(envs.ServiceURL) > 1 {
		herroku.RunKeepAlive()
	} else {
		log.Println("[INFO] ServiceURL is empty, heroku keep-alive is not used")
	}

	store, err := makeDataStore(envs)
	if err != nil {
		log.Fatalf("[FATAL] failed to make data store, %s", err)
	}

	app := App{
		envs:  envs,
		store: store,
	}

	_ = app

}

func makeDataStore(envs *AppEnvs) (storeReturned store.Store, err error) {
	switch envs.StoreType {
	case "mongo":
		{
			storeReturned, err = store.NewMongoDB(envs.MongoURL, envs.MongoDB)
		}
	}
	return
}
