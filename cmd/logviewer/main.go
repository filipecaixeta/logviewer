package main

import (
	"fmt"
	"log"
	"os"

	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/model"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version    = "dev"
	Cfg        config.Config
	cfgFile    string
	namespaces []string
	context    string
	flagL      bool
	flagD      bool
)

// rootCmd represents the root command for the LogViewer TUI tool.
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "LogViewer TUI",
	Long: `This tool can be used to view logs from Kubernetes or Docker containers.. 

Features:
- Interactive List Navigation: Browse Kubernetes namespaces, workloads, pods, and containers using keyboard controls.
- Real-time Logs Viewing: Stream logs from selected Docker or Kubernetes containers.
- Log Filtering and Manipulation: Customize namespaces to display via a TOML configuration file.
`,
}

func commonRunE(commandName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Set the command in config
		Cfg.Command = commandName
		var p *tea.Program

		// if any flag is set replace the config value
		Cfg.Color = "dark"
		if cmd.Flags().Changed("namespaces") {
			Cfg.Namespaces = namespaces
		}
		if cmd.Flags().Changed("context") {
			Cfg.K8sContext = context
		}
		if cmd.Flags().Changed("light") {
			Cfg.Color = "light"
		}
		if cmd.Flags().Changed("dark") {
			Cfg.Color = "dark"
		}
		if cmd.Flags().Changed("light") && cmd.Flags().Changed("dark") {
			fmt.Println("Error: --light and --dark are mutually exclusive")
			os.Exit(1)
		}

		var logFile *os.File
		if os.Getenv("DEBUG") == "true" {
			logFile, _ = os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		} else {
			logFile, _ = os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		}
		oldStdout := os.Stdout
		os.Stdout = logFile
		defer func() { os.Stdout = oldStdout }()
		defer logFile.Close()
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic:", r)
				// Exit Bubble Tea program gracefully
				if p != nil {
					p.Quit()
				}
				// Exit the application with a non-zero status code
				os.Exit(1)
			}
		}()

		config.SetColor(Cfg.Color)

		m := model.New(cmd.Context(), &Cfg)

		p = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
		if _, err := p.Run(); err != nil {
			os.Stdout = oldStdout
			logFile.Close()
			log.Fatalf("Error starting TUI: %s\n", err)
		}

		if m.Err != nil {
			log.Fatalf("Error: %s\n", m.Err)
		}

		return nil
	}
}

func newK8sCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes Command",
		Long:  `Interact with Kubernetes clusters.`,
		RunE:  commonRunE("k8s"),
	}
	c.PersistentFlags().StringSliceVarP(&namespaces, "namespaces", "n", []string{}, "namespaces to use")
	c.PersistentFlags().StringVarP(&context, "context", "c", "", "k8s context to use")
	return c
}

func newStdinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stdin",
		Short: "Standard Input Command",
		Long:  `Read from standard input.`,
		RunE:  commonRunE("stdin"),
	}
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Standard Input Command",
		Long:  `Read from standard input.`,
		RunE:  commonRunE("test"),
	}
}

func newDockerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "docker",
		Short: "Docker Command",
		Long:  `Interact with Docker containers.`,
		RunE:  commonRunE("docker"),
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:", Version)
		},
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.toml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&flagL, "light", "l", false, "Use the light mode")
	rootCmd.PersistentFlags().BoolVarP(&flagD, "dark", "d", true, "Use the dark mode (default)")
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Println("Unable to bind flags:", err)
	}

	rootCmd.AddCommand(newK8sCmd())
	rootCmd.AddCommand(newStdinCmd())
	rootCmd.AddCommand(newDockerCmd())
	rootCmd.AddCommand(newTestCmd())
	rootCmd.AddCommand(newVersionCmd())
}

// initConfig loads the config from the toml config file and unmarshals
// it into Cfg. Command line flags override the config file.
func initConfig() {
	if cfgFile == "" {
		cfgFile = findConfigFile()
	}

	// try to open the file and read it
	_, err := os.Stat(cfgFile)
	if err != nil {
		fmt.Println("Error opening config file:", err)
	}

	// read into a string
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}

	// unmarshal toml into Cfg
	if _, err := toml.Decode(string(b), &Cfg); err != nil {
		fmt.Println("Error decoding config file:", err)
	}

	Cfg.Filename = cfgFile
}

// findConfigFile searches for a config file in the following order
// 1. env var LOGVIEWER_CONFIG
// 2. the current working directory
// 3. the user's home directory
func findConfigFile() string {
	// 1. env var LOGVIEWER_CONFIG
	if c := os.Getenv("LOGVIEWER_CONFIG"); c != "" {
		return c
	}

	// 2. the current working directory
	if _, err := os.Stat("config.toml"); err == nil {
		wd, err := os.Getwd()
		if err != nil {
			return "config.toml"
		}
		return wd + "/config.toml"
	}

	// 3. the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}
	return home + "/config.toml"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
