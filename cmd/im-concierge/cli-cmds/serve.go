package cmd

import (
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/flashmob/go-guerrilla"
	imc "github.com/sapiens-sapide/IM-concierge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"os"
	"os/signal"
	"syscall"
)

var (
	configPath    string
	configFile    string
	pidFile       string
	signalChannel chan os.Signal // for trapping SIG_HUP
	cmdConfig     imc.ConciergeConfig

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "starts IM-concierge server",
		Run:   serve,
	}
)

func init() {
	serveCmd.PersistentFlags().StringVarP(&configFile, "config", "c",
		"im-concierge-config", "Name of the configuration file, without extension. (YAML, TOML, JSON… allowed)")
	serveCmd.PersistentFlags().StringVarP(&configPath, "configpath", "",
		"../../", "Main config file path.")
	serveCmd.PersistentFlags().StringVarP(&pidFile, "pid-file", "p",
		"/var/run/IM-concierge.pid", "Path to the pid file")

	RootCmd.AddCommand(serveCmd)
	signalChannel = make(chan os.Signal, 1)
	cmdConfig = imc.ConciergeConfig{}
}

func sigHandler() {
	// handle SIGHUP for reloading the configuration while running
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)

	for sig := range signalChannel {

		if sig == syscall.SIGHUP {
			err := readConfig(&cmdConfig)
			if err != nil {
				log.WithError(err).Error("Error while ReadConfig (reload)")
			} else {
				log.Infof("Configuration is reloaded at %s", guerrilla.ConfigLoadTime)
			}
			// TODO: reinitialize
		} else if sig == syscall.SIGTERM || sig == syscall.SIGQUIT || sig == syscall.SIGINT {
			log.Infof("Shutdown signal caught")
			//TODO: graceful shutdown
			log.Infof("Shutdown completed, exiting.")
			os.Exit(0)
		} else {
			os.Exit(0)
		}
	}
}

func serve(cmd *cobra.Command, args []string) {
	err := readConfig(&cmdConfig)
	if err != nil {
		log.WithError(err).Fatal("Error while reading config")
	}
	imc.Config = cmdConfig
	imc.Users = make(map[string]imc.Recipient)
	// Write out our PID
	if len(pidFile) > 0 {
		if f, err := os.Create(pidFile); err == nil {
			defer f.Close()
			if _, err := f.WriteString(fmt.Sprintf("%d", os.Getpid())); err == nil {
				f.Sync()
			} else {
				log.WithError(err).Fatalf("Error while writing pidFile (%s)", pidFile)
			}
		} else {
			log.WithError(err).Fatalf("Error while creating pidFile (%s)", pidFile)
		}
	}

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	if err != nil {
		log.WithError(err).Fatalln("cayley.NewGraph error")
	}

	irccon := irc.IRC(cmdConfig.IRCNickname, cmdConfig.IRCUser)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(cmdConfig.IRCChannel)
	})
	irccon.AddCallback("353", func(e *irc.Event) {
		imc.HandleUsersList(irccon, e)
	})
	irccon.AddCallback("JOIN", func(e *irc.Event) {
		imc.HandleUsersList(irccon, e)
	})
	irccon.AddCallback("352", imc.HandleWhoReply)
	irccon.AddCallback("PRIVMSG", imc.HandleMessage)
	err = irccon.Connect(cmdConfig.IRCserver)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.GetNick()

	go irccon.Loop()
	sigHandler()
}

// ReadConfig which should be called at startup, or when a SIG_HUP is caught
func readConfig(config *imc.ConciergeConfig) error {
	// load in the main config. Reading from YAML, TOML, JSON, HCL and Java properties config files
	v := viper.New()
	v.SetConfigName(configFile) // name of config file (without extension)
	v.AddConfigPath(configPath) // path to look for the config file in
	v.AddConfigPath("../../")   // call multiple times to add many search paths
	v.AddConfigPath(".")        // optionally look for config in the working directory

	err := v.ReadInConfig() // Find and read the config file*/
	if err != nil {
		log.WithError(err).Infof("Could not read main config file <%s>.", configFile)
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		log.WithError(err).Infof("Could not parse config file: <%s>", configFile)
		return err
	}

	return nil
}
