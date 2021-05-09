package main

import (
	"banner_rotation/internal/app"
	"banner_rotation/internal/server"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	sqlrepository "banner_rotation/internal/repository/sql"

	_ "banner_rotation/migrations"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pressly/goose"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const (
	driver        = "postgres"
	migrationsDir = "migrations"
	layoutTime    = "01-02-2006-15-04-05"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "banner_rotation",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.banner_rotation.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		// Search config in home directory with name "calendar.json".
		viper.AddConfigPath(home)
		viper.SetConfigName("banners.json")
	}

	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.BindEnv("dblogin", "POSTGRES_USER")
	viper.BindEnv("dbname", "POSTGRES_DB")
	viper.BindEnv("dbpassword", "POSTGRES_PASSWORD")

	readConfig()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config was changed: ", e.Name)
		readConfig()
	})
}

func readConfig() {
	if err := viper.ReadInConfig(); err == nil {
		log.Info("using config file: "+viper.ConfigFileUsed(), log.Fields{"settings": viper.AllSettings()})
	} else {
		log.Fatal("failed to read config file: " + err.Error())
	}
}

func run(_ *cobra.Command, args []string) {
	config, err := NewConfig()
	if err != nil {
		log.Fatal("failed to read config: ", err.Error())
	}

	logF, err := configureLogger(config)
	if err != nil {
		log.Fatal("failed to configure logger: ", err.Error())
	}
	defer func() {
		if err := logF.Close(); err != nil {
			log.Panic("failed to close log file: ", err.Error())
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sqlx.Connect(driver, fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		config.DataBase.Host, config.DataBase.Port, config.DataBase.Login, config.DataBase.DBName, config.DataBase.Password))
	if err != nil {
		log.WithFields(log.Fields{
			"host":     config.DataBase.Host,
			"port":     config.DataBase.Port,
			"login":    config.DataBase.Login,
			"dbname":   config.DataBase.DBName,
			"password": config.DataBase.Password,
		}).Fatal("failed to connect to database :", err.Error())
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal("failed to close database: ", err.Error())
		}
	}()

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		log.Error("failed to migrate: ", err.Error())
		return
	}

	repo := sqlrepository.NewSQLRepository(db.DB)
	a := app.NewBannersApp(repo)
	s := server.NewServer(a, config.Server.Port, config.Server.Host)

	if err := s.Start(ctx); err != nil {
		log.Error("failed to start http server: " + err.Error())
		cancel()
		return
	}
	log.Info("server was started...")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)

		select {
		case <-ctx.Done():
			return
		case <-signals:
		}

		signal.Stop(signals)
		cancel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := s.Stop(ctx); err != nil {
			log.Error("failed to stop http server: " + err.Error())
		} else {
			log.Info("server was stopped")
		}
	}()

	wg.Wait()
}

func configureLogger(c Config) (*os.File, error) {
	l, err := log.ParseLevel(c.Logger.Level)
	if err != nil {
		log.WithField("level string", c.Logger.Level).Errorf("failed to parse level: " + err.Error())
		return nil, err
	}
	log.SetLevel(l)

	fileName := fmt.Sprint("banners", time.Now().Format(layoutTime), ".log")
	f, err := os.OpenFile(path.Join(c.Logger.Path, fileName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 755)
	if err != nil {
		log.WithFields(log.Fields{
			"path":      c.Logger.Path,
			"file name": fileName,
		}).Errorf("failed to create log file: " + err.Error())
		return nil, err
	}
	log.SetOutput(f)

	return f, nil
}
