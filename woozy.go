package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"sort"

	"github.com/gummiboll/woozy/weatherdata"
)

// Configuration for woozy
type Configuration struct {
	Place string `json:"place"`
	Days  int    `json:"days"`
}

var (
	icons = map[string]string{
		"Rain":               "\u2602",
		"Light rain":         "\u2602",
		"Light rain showers": "\u2602",
		"Heavy rain":         "\u2602",
		"Rain showers":       "\u2602",
		"Clear sky":          "\u263C",
		"Cloudy":             "\u2601",
		"Partly cloudy":      "\u26C5",
		"Fair":               "\u26C5",
	}

	windIcons = map[string]string{
		"N":   "\u2193",
		"S":   "\u2191",
		"E":   "\u2190",
		"W":   "\u2192",
		"NE":  "\u2199",
		"NNE": "\u2199",
		"NEE": "\u2199",
		"ENE": "\u2199",
		"NW":  "\u2198",
		"NNW": "\u2198",
		"NWW": "\u2198",
		"WNW": "\u2198",
		"SE":  "\u2196",
		"SEE": "\u2196",
		"SSE": "\u2196",
		"ESE": "\u2196",
		"SW":  "\u2197",
		"SSW": "\u2197",
		"SWW": "\u2197",
		"WSW": "\u2197",
	}
)

func printHeader(wd weatherdata.WeatherData) {
	hStr := fmt.Sprintf("%s / %s", wd.Location.Country, wd.Location.Name)
	hStr = fmt.Sprintf("%s | %s %s %s %s %s (%.1f hours)", hStr, icons["Clear sky"], windIcons["S"], wd.Sun.Rise.Local().Format("15:04"), windIcons["N"], wd.Sun.Set.Local().Format("15:04"), wd.SunHours())
	fmt.Println(hStr)
}

func printFooter(wd weatherdata.WeatherData) {
	fmt.Println(fmt.Sprintf("Forecast issued %.1f hours ago, next update in %.1f hours", wd.Meta.HoursSinceUpdate(), wd.Meta.HoursToNextUpdate()))
	fmt.Println(fmt.Sprintf("\n%s.\n%s", wd.Credit.Text, wd.Credit.URL))
}

func printForecast(wd weatherdata.WeatherData, days int) {
	forecasts := make(map[string][]weatherdata.WeatherForecast)
	for i := range wd.Forecast {
		forecasts[wd.Forecast[i].From.Format("02/01")] = append(forecasts[wd.Forecast[i].From.Format("02/01")], wd.Forecast[i])
	}

	mk := make([]string, len(forecasts))
	i := 0
	for k := range forecasts {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	for i := 0; i < days; i++ {
		var wStr string
		fmt.Println(forecasts[mk[i]][0].From.Format("January 02 (Monday):"))
		for _, w := range forecasts[mk[i]] {
			wStr = fmt.Sprintf(" %s %s %d%s (%.1fmm)", w.PeriodName(), icons[w.Symbol.Name], w.Temperature.Value, "\u2103", w.Precipitation.Value)
			wStr = fmt.Sprintf("%s %s %.1f m/s", wStr, windIcons[w.WindDirection.Code], w.WindSpeed.Mps)
			fmt.Println(wStr)
		}
		fmt.Println("")
	}
}

func loadConfig() (Configuration, error) {
	configuration := Configuration{}

	usr, err := user.Current()
	if err != nil {
		return configuration, err
	}

	cfile := path.Join(usr.HomeDir, ".woozy")

	file, err := os.Open(cfile)
	if err != nil {
		fmt.Println(fmt.Sprintf("Configuration file not found, creating a example in %s..", cfile))
		out, err := os.Create(cfile)
		if err != nil {
			panic(err)
		}
		defer out.Close()
		configuration.Place = "Sweden/Västerbotten/Estersmark"
		configuration.Days = 3
		configjson, err := json.MarshalIndent(configuration, "", "\t")
		ioutil.WriteFile(cfile, configjson, 0600)
		fmt.Println(".. edit it and restart woozy")
		os.Exit(1)
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)

	if err != nil {
		return configuration, err
	}

	return configuration, nil

}

func main() {
	configuration, err := loadConfig()

	if err != nil {
		panic(err)
	}

	cc := flag.Bool("cache-clear", false, "Force cache clear")
	days := flag.Int("days", 3, "Number of days to print")
	flag.Parse()

	wd, err := weatherdata.LoadWeatherData(configuration.Place, *cc)
	if err != nil {
		panic(err)
	}

	if configuration.Days == 0 {
		configuration.Days = *days
	}

	printHeader(wd)
	printForecast(wd, configuration.Days)
	printFooter(wd)

}