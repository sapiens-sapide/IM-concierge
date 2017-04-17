package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/sapiens-sapide/IM-concierge/concierge"
	ent "github.com/sapiens-sapide/IM-concierge/entities"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var (
	configPath    string
	configFile    string
	pidFile       string
	signalChannel chan os.Signal // for trapping SIG_HUP
	cmdConfig     ent.ConciergeConfig

	serveCmd = &cobra.Command{
		Use:   "run",
		Short: "starts IM-concierge server",
		Run:   run,
	}
)

const version = "0.2.0"

func init() {
	serveCmd.PersistentFlags().StringVarP(&configFile, "config", "c",
		"im-concierge-config", "Name of the configuration file, without extension. (YAML, TOML, JSON… allowed)")
	serveCmd.PersistentFlags().StringVarP(&configPath, "configpath", "",
		"../../", "Main config file path.")
	serveCmd.PersistentFlags().StringVarP(&pidFile, "pid-file", "p",
		"/var/run/IM-concierge.pid", "Path to the pid file")

	RootCmd.AddCommand(serveCmd)
	signalChannel = make(chan os.Signal, 1)
	cmdConfig = ent.ConciergeConfig{}
}

func sigHandler(c *concierge.Concierge) {
	// handle SIGHUP for reloading the configuration while running
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)

	for sig := range signalChannel {

		if sig == syscall.SIGHUP {
			err := readConfig(&cmdConfig)
			if err != nil {
				log.WithError(err).Error("Error while ReadConfig (reload)")
			} else {
				//TODO
			}
			// TODO: reinitialize
		} else if sig == syscall.SIGTERM || sig == syscall.SIGQUIT || sig == syscall.SIGINT {
			log.Infof("Shutdown signal caught")
			err := c.Shutdown()
			if err != nil {
				log.WithError(err).Warnf("error when shutingdown backend")
			}
			log.Infof("Shutdown completed, exiting.")
			os.Exit(0)
		} else {
			os.Exit(0)
		}
	}
}

func run(cmd *cobra.Command, args []string) {

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

	// Load configuration
	err := readConfig(&cmdConfig)
	if err != nil {
		log.WithError(err).Fatal("Error while reading config")
	}

	// Create a concierge
	concierge, err := concierge.NewConcierge(cmdConfig)
	if err != nil {
		log.WithError(err).Fatal("unable to create a concierge")
	}

	// Start it
	err = concierge.Start()
	if err != nil {
		log.WithError(err).Fatal("unable to start concierge")
	}

	//wait for system signals
	sigHandler(concierge)
}

// ReadConfig which should be called at startup, or when a SIG_HUP is caught
func readConfig(config *ent.ConciergeConfig) error {
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
