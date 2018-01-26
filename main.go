package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/juju/loggo"

	"github.com/godbus/dbus"
	"github.com/gosexy/gettext"
	"github.com/gotk3/gotk3/gtk"
	"github.com/okelet/goutils"
	"github.com/okelet/proxychanger/proxychangerlib"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var Version = "master"

func main() {

	var err error

	err = proxychangerlib.InitConstants()
	if err != nil {
		fmt.Printf("Error during initialization: %v\n", err)
		goutils.ShowZenityError("Error", fmt.Sprintf("Error initializing application: %v.", err))
		os.Exit(1)
	}

	gettext.BindTextdomain(proxychangerlib.GETTEXT_DOMAIN, proxychangerlib.LOCALE_DIR)
	gettext.Textdomain(proxychangerlib.GETTEXT_DOMAIN)
	gettext.SetLocale(gettext.LcAll, "")

	sessionBus, err := dbus.SessionBus()
	if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error getting dbus session connection: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error getting dbus session connection: %v.", err))
		os.Exit(1)
	}

	app := kingpin.New("proxychanger", proxychangerlib.MyGettextv("Changes proxy settings in multiple applications")).Version(Version)
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	logLevel := app.Flag("log-level", proxychangerlib.MyGettextv("Log level")).Short('l').String()
	logErrorOutput := app.Flag("error-output", proxychangerlib.MyGettextv("Log to error output")).Bool()

	indicatorCommand := app.Command("indicator", proxychangerlib.MyGettextv("Start as an indicator"))
	configFile := indicatorCommand.Flag("config", proxychangerlib.MyGettextv("Configuration file")).Short('c').ExistingFile()
	testMode := indicatorCommand.Flag("test", proxychangerlib.MyGettextv("Test mode")).Short('t').Bool()
	indicatorCommand.Default()

	listCommand := app.Command("list", proxychangerlib.MyGettextv("List proxies"))

	setActiveCommand := app.Command("set", proxychangerlib.MyGettextv("Set active proxy"))
	setActiveCommandSlug := setActiveCommand.Arg("slug", proxychangerlib.MyGettextv("New active proxy slug; use 'none' to unset the proxy")).Required().String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	cmdLogLevelSet := false
	if *logLevel != "" {
		level, ok := loggo.ParseLevel(*logLevel)
		if !ok {
			fmt.Println(proxychangerlib.MyGettextv("Invalid LOG level: %v.", *logLevel))
			goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Invalid LOG level: %v.", *logLevel))
			os.Exit(1)
		}
		proxychangerlib.Log.SetLogLevel(level)
		cmdLogLevelSet = true
	} else {
		// TODO: Change log level according to config
		// proxychangerlib.Log.SetLogLevel(level)
	}

	if logErrorOutput != nil && *logErrorOutput {
		proxychangerlib.AddErrorOutputLogging()
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case indicatorCommand.FullCommand():
		os.Exit(runIndicator(sessionBus, *configFile, cmdLogLevelSet, *testMode))
	case listCommand.FullCommand():
		listProxies(sessionBus)
	case setActiveCommand.FullCommand():
		setActiveProxyBySlug(sessionBus, *setActiveCommandSlug)
	}

}

func runIndicator(sessionBus *dbus.Conn, configFile string, cmdLogLevelSet bool, testMode bool) int {

	reply, err := sessionBus.RequestName(proxychangerlib.DBUS_INTERFACE, dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error requesting dbus name: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error requesting dbus name: %v.", err.Error()))
		return 1
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Println(proxychangerlib.MyGettextv("Program already running."))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Program already running."))
		return 1
	}

	config, warnings, err := proxychangerlib.NewConfig(sessionBus, configFile, false)
	if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error loading configuration: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error loading configuration: %v.", err))
		return 1
	}
	if len(warnings) > 0 {
		// TODO Log
	}

	proxychangerlib.Log.SetLogLevel(config.LogLevel)

	gtk.Init(nil)

	i, err := proxychangerlib.NewIndicator(sessionBus, config, Version, cmdLogLevelSet, testMode)
	if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error creating indicator: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error creating indicator: %v.", err))
		return 1
	}
	err = i.Run(true)
	if err == nil {
		gtk.Main()
		return 0
	} else {
		fmt.Println("Error running indicator", err)
		return 1
	}

}

func getConfigConnection(dbusConnection *dbus.Conn) (proxychangerlib.ConfigInterface, error) {
	return proxychangerlib.NewDbusConfig(dbusConnection)
}

func listProxies(dbusConnection *dbus.Conn) int {

	c, err := getConfigConnection(dbusConnection)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	responseData, err := c.DbusListProxies()

	var response proxychangerlib.DbusListProxiesResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		proxychangerlib.MyGettextv("Active"),
		proxychangerlib.MyGettextv("Slug"),
		proxychangerlib.MyGettextv("Name"),
		proxychangerlib.MyGettextv("Protocol"),
		proxychangerlib.MyGettextv("Address"),
		proxychangerlib.MyGettextv("Username"),
	})

	for _, v := range response.Proxies {
		table.Append([]string{
			map[bool]string{true: proxychangerlib.MyGettextv("Yes"), false: proxychangerlib.MyGettextv("No")}[v.Active],
			v.Slug,
			v.Name,
			v.Protocol,
			v.Address,
			v.Username,
		})
	}
	table.Render() // Send output

	return 0

}

func setActiveProxyBySlug(dbusConnection *dbus.Conn, slug string) int {

	c, err := getConfigConnection(dbusConnection)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	responseData, err := c.SetActiveProxyBySlug(slug)

	var response proxychangerlib.DbusSetActiveProxyBySlugResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		proxychangerlib.MyGettextv("Error setting active proxy: %v", response.Error)
		return 1
	}

	return 0

}
