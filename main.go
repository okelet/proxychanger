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
	configFile := app.Flag("config", proxychangerlib.MyGettextv("Configuration file")).Short('c').ExistingFile()

	indicatorCommand := app.Command("indicator", proxychangerlib.MyGettextv("Start as an indicator"))
	testMode := indicatorCommand.Flag("test", proxychangerlib.MyGettextv("Test mode")).Short('t').Bool()
	indicatorCommand.Default()

	listCommand := app.Command("list", proxychangerlib.MyGettextv("List proxies"))
	// TODO: output format
	includePasswords := listCommand.Flag("include-passwords", proxychangerlib.MyGettextv("Include passwords")).Bool()

	applyActiveCommand := app.Command("apply", proxychangerlib.MyGettextv("Apply current current active proxy"))

	getActiveCommand := app.Command("get", proxychangerlib.MyGettextv("Get current active proxy slug; returns empty if no active proxy"))

	setActiveCommand := app.Command("set", proxychangerlib.MyGettextv("Set active proxy"))
	setActiveCommandSlug := setActiveCommand.Arg("slug", proxychangerlib.MyGettextv("New active proxy slug; use 'none' to unset the proxy")).Required().String()

	// TODO: add command

	// TODO: edit command

	// TODO: delete command

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
	}

	if *logErrorOutput {
		proxychangerlib.AddErrorOutputLogging()
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case indicatorCommand.FullCommand():
		os.Exit(runIndicator(sessionBus, *configFile, cmdLogLevelSet, *testMode))
	case listCommand.FullCommand():
		listProxies(sessionBus, *configFile, cmdLogLevelSet, *includePasswords)
	case applyActiveCommand.FullCommand():
		applyActiveProxyBySlug(sessionBus, *configFile, cmdLogLevelSet)
	case getActiveCommand.FullCommand():
		getActiveProxyBySlug(sessionBus, *configFile, cmdLogLevelSet)
	case setActiveCommand.FullCommand():
		setActiveProxyBySlug(sessionBus, *setActiveCommandSlug, *configFile, cmdLogLevelSet)
	}

}

func runIndicator(sessionBus *dbus.Conn, configFile string, cmdLogLevelSet bool, testMode bool) int {

	config, warnings, err := proxychangerlib.NewConfig(configFile, false)
	if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error loading configuration: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error loading configuration: %v.", err))
		return 1
	}

	err = config.Lock(sessionBus)
	if err == proxychangerlib.ApplicationAlreadyRunningError {
		fmt.Println(proxychangerlib.MyGettextv("Application already running."))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Application already running."))
		return 1
	} else if err != nil {
		fmt.Println(proxychangerlib.MyGettextv("Error locking configuration: %v.", err))
		goutils.ShowZenityError(proxychangerlib.MyGettextv("Error"), proxychangerlib.MyGettextv("Error locking configuration: %v.", err))
		return 1
	}

	if len(warnings) > 0 {
		// TODO Log
	}

	if !cmdLogLevelSet {
		proxychangerlib.Log.SetLogLevel(config.LogLevel)
	}

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

func getConfigService(dbusConnection *dbus.Conn, configFile string, cmdLogLevelSet bool) (proxychangerlib.ConfigService, []string, error) {

	// Get the current list of names to check if already running
	var names []string
	err := dbusConnection.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return nil, nil, err
	}

	if goutils.ListContainsString(names, proxychangerlib.DBUS_INTERFACE) {
		// Already running, connect using dbus
		return proxychangerlib.NewConfigDbus(dbusConnection)
	} else {

		// Not running, create a new configuration and lock it
		config, warnings, err := proxychangerlib.NewConfig(configFile, false)
		if err != nil {
			return nil, nil, err
		}

		err = config.Lock(dbusConnection)
		if err != nil {
			return nil, nil, err
		}

		if !cmdLogLevelSet {
			proxychangerlib.Log.SetLogLevel(config.LogLevel)
		}

		return config, warnings, nil

	}

}

func listProxies(dbusConnection *dbus.Conn, configFile string, cmdLogLevelSet bool, includePasswords bool) int {

	c, warnings, err := getConfigService(dbusConnection, configFile, cmdLogLevelSet)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	if len(warnings) > 0 {
		// TODO
	}

	responseData, err := c.ListProxies(includePasswords)

	var response proxychangerlib.ListProxiesResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		fmt.Println(proxychangerlib.MyGettextv("Error listing proxies: %v.", response.Error))
		return 1
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{
		proxychangerlib.MyGettextv("Active"),
		proxychangerlib.MyGettextv("Slug"),
		proxychangerlib.MyGettextv("Name"),
		proxychangerlib.MyGettextv("Protocol"),
		proxychangerlib.MyGettextv("Address"),
		proxychangerlib.MyGettextv("Username"),
	}
	if includePasswords {
		header = append(header, proxychangerlib.MyGettextv("Password"))
	}
	table.SetHeader(header)

	for _, v := range response.Proxies {
		row := []string{
			map[bool]string{true: proxychangerlib.MyGettextv("Yes"), false: proxychangerlib.MyGettextv("No")}[v.Active],
			v.Slug,
			v.Name,
			v.Protocol,
			v.Address,
			v.Username,
		}
		if includePasswords {
			row = append(row, v.Password)
		}
		table.Append(row)
	}
	table.Render()

	return 0

}

func applyActiveProxyBySlug(dbusConnection *dbus.Conn, configFile string, cmdLogLevelSet bool) int {

	c, warnings, err := getConfigService(dbusConnection, configFile, cmdLogLevelSet)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	if len(warnings) > 0 {
		// TODO
	}

	responseData, err := c.ApplyActiveProxy()

	var response proxychangerlib.ApplyActiveProxyResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		fmt.Println(proxychangerlib.MyGettextv("Error applying active proxy: %v.", response.Error))
		return 1
	}

	return 0

}

func getActiveProxyBySlug(dbusConnection *dbus.Conn, configFile string, cmdLogLevelSet bool) int {

	c, warnings, err := getConfigService(dbusConnection, configFile, cmdLogLevelSet)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	if len(warnings) > 0 {
		// TODO
	}

	responseData, err := c.GetActiveProxySlug()

	var response proxychangerlib.GetActiveProxySlugResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		fmt.Println(proxychangerlib.MyGettextv("Error setting active proxy: %v.", response.Error))
		return 1
	} else {
		fmt.Println(response.Slug)
	}

	return 0

}

func setActiveProxyBySlug(dbusConnection *dbus.Conn, slug string, configFile string, cmdLogLevelSet bool) int {

	c, warnings, err := getConfigService(dbusConnection, configFile, cmdLogLevelSet)
	if err != nil {
		fmt.Println("Error", err)
		return 1
	}

	if len(warnings) > 0 {
		// TODO
	}

	responseData, err := c.SetActiveProxyBySlug(slug)

	var response proxychangerlib.SetActiveProxyBySlugResponse
	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		fmt.Println(proxychangerlib.MyGettextv("Error setting active proxy: %v.", response.Error))
		return 1
	}

	return 0

}
