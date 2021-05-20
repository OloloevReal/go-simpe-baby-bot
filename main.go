package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OloloevReal/go-simple-baby-bot/goenvs"
	"github.com/OloloevReal/go-simple-baby-bot/heroku"
	"github.com/OloloevReal/go-simple-baby-bot/store"
	"github.com/OloloevReal/go-simple-baby-bot/telegram"
	log "github.com/OloloevReal/go-simple-log"
	"golang.org/x/net/proxy"
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
	ctx   context.Context
	bot   telegram.TelegramBot
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

	var prxy proxy.Dialer
	if envs.TelegramProxy != "" {
		prxy, err = proxy.SOCKS5("tcp", envs.TelegramProxy, nil, proxy.Direct)
		if err != nil {
			log.Fatalf("[FATAL] failed to set proxy %s", envs.TelegramProxy)
		}
	}

	bot, err := telegram.NewTelegram(&telegram.TelegramConfig{Token: envs.TelegramToken,
		Proxy:   prxy,
		Debug:   envs.Debug,
		Timeout: time.Second * 7,
	})
	if err != nil {
		log.Fatalf("[FATAL] failed to make a telegram bot")
	}

	app := App{
		envs:  envs,
		store: store,
		ctx:   ctx,
		bot:   *bot,
	}

	app.Start()

}

func (a *App) Start() {
	a.bot.Run(a.ctx)
	// log.Println("Waiting to finish")
	// <-a.ctx.Done()
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
