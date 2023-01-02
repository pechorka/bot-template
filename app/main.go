package main

import (
	"context"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jessevdk/go-flags"
	"github.com/pechorka/bot-template/app/bot"
	"github.com/pechorka/bot-template/app/events"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

var opts struct {
	Telegram struct {
		Token string `long:"token" env:"TOKEN" description:"telegram bot token" required:"true"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Dbg bool `long:"dbg" env:"DEBUG" description:"debug mode"`
}

func exportEnvFile() error {
	f, err := os.ReadFile(".env")
	if err != nil {
		return errors.Wrap(err, "failed to read .env file")
	}

	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		name, val, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New("failed to parse .env file")
		}
		if err := os.Setenv(name, val); err != nil {
			return errors.Wrap(err, "failed to set env var")
		}
	}

	return nil
}

func main() {
	ctx := context.TODO()

	if err := exportEnvFile(); err != nil {
		log.Fatalf("[WARN] failed to export env vars from .env file: %v", err)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatalf("[ERROR] failed to parse flags: %v", err)
	}
	tbAPI, err := tbapi.NewBotAPI(opts.Telegram.Token)
	if err != nil {
		log.Fatalf("[ERROR] can't make telegram bot, %v", err)
	}
	tbAPI.Debug = opts.Dbg

	tgListener := events.Listener{
		API: tbAPI,
		Bot: &bot.Echo{},
	}

	if err := tgListener.Do(ctx); err != nil {
		log.Fatalf("[ERROR] telegram listener failed, %v", err)
	}
}
