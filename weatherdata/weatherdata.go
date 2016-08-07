package weatherdata

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	cacheFile  = fmt.Sprintf("%swoozyforecast.xml", os.TempDir())
	dateFormat = "2006-01-02T15:04:05"
)

type weatherCredit struct {
	URL  string `xml:"url,attr"`
	Text string `xml:"text,attr"`
}

type customTime struct {
	time.Time
}

type customTimeAttr struct {
	time.Time
}

type sun struct {
	Rise customTimeAttr `xml:"rise,attr"`
	Set  customTimeAttr `xml:"set,attr"`
}

type weatherLocation struct {
	Name     string `xml:"name"`
	Type     string `xml:"type"`
	Country  string `xml:"country"`
	Timezone struct {
		ID        string `xml:"id,attr"`
		UTCOffset string `xml:"utcoffsetMinutes,attr"`
	} `xml:"timezone"`
}

// WeatherMeta contains metadata about the forecast
type WeatherMeta struct {
	LastUpdate customTime `xml:"lastupdate"`
	NextUpdate customTime `xml:"nextupdate"`
}

// WeatherForecast contains actual forecast
type WeatherForecast struct {
	From     customTimeAttr `xml:"from,attr"`
	To       customTimeAttr `xml:"to,attr"`
	Period   int            `xml:"period,attr"`
	Pressure struct {
		Unit  string  `xml:"unit,attr"`
		Value float32 `xml:"value,attr"`
	} `xml:"pressure"`
	Precipitation struct {
		Value float32 `xml:"value,attr"`
		Min   float32 `xml:"minvalue,attr"`
		Max   float32 `xml:"maxvalue,attr"`
	} `xml:"precipitation"`
	Symbol struct {
		Name   string `xml:"name,attr"`
		Number int    `xml:"number,attr"`
	} `xml:"symbol"`
	Temperature struct {
		Unit  string `xml:"unit,attr"`
		Value int    `xml:"value,attr"`
	} `xml:"temperature"`
	WindDirection struct {
		Deg  float32 `xml:"deg,attr"`
		Code string  `xml:"code,attr"`
		Name string  `xml:"name,attr"`
	} `xml:"windDirection"`
	WindSpeed struct {
		Mps  float32 `xml:"mps,attr"`
		Name string  `xml:"name,attr"`
	} `xml:"windSpeed"`
}

// WeatherData contains actual weather data
type WeatherData struct {
	Credit   weatherCredit     `xml:"credit>link"`
	Location weatherLocation   `xml:"location"`
	Meta     WeatherMeta       `xml:"meta"`
	Sun      sun               `xml:"sun"`
	Forecast []WeatherForecast `xml:"forecast>tabular>time"`
}

func (c *customTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	loc, _ := time.LoadLocation("Local")
	parse, err := time.ParseInLocation(dateFormat, v, loc)
	if err != nil {
		return err
	}
	*c = customTime{parse}
	return nil
}

func (ca *customTimeAttr) UnmarshalXMLAttr(attr xml.Attr) error {
	loc, _ := time.LoadLocation("Local")
	parse, err := time.ParseInLocation(dateFormat, attr.Value, loc)
	if err != nil {
		return err
	}
	*ca = customTimeAttr{parse}
	return nil
}

// PeriodName returns period as text
func (wf WeatherForecast) PeriodName() string {
	var retStr string
	switch {
	case wf.Period == 0:
		retStr = "Night:   "
	case wf.Period == 1:
		retStr = "Morning: "
	case wf.Period == 2:
		retStr = "Day:     "
	case wf.Period == 3:
		retStr = "Evening: "
	}
	return retStr
}

// HoursSinceUpdate returns hours since last update
func (wm WeatherMeta) HoursSinceUpdate() float64 {
	t := time.Unix(0, wm.LastUpdate.Local().UnixNano())
	elapsed := time.Since(t)
	return elapsed.Hours()
}

// HoursToNextUpdate returns hours until next update
func (wm WeatherMeta) HoursToNextUpdate() float64 {
	t := time.Unix(0, wm.NextUpdate.Local().UnixNano())
	elapsed := t.Sub(time.Now())
	return elapsed.Hours()
}

// SunHours represents number of hours sun is up
func (wd WeatherData) SunHours() float64 {
	t := time.Unix(0, wd.Sun.Set.Local().UnixNano())
	elapsed := t.Sub(time.Unix(0, wd.Sun.Rise.Local().UnixNano()))
	return elapsed.Hours()
}

func yrURL(place string) string {
	return fmt.Sprintf("http://www.yr.no/place/%s/forecast.xml", place)
}

func weatherDataCache(cc bool, assumeValid bool) (wd WeatherData, valid bool) {
	if cc == true {
		fmt.Println("Clearing cache")
		os.Remove(cacheFile)
		return wd, false
	}
	xmlFile, err := os.Open(cacheFile)

	if err != nil {
		return wd, false
	}

	defer xmlFile.Close()
	XMLdata, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(XMLdata, &wd)

	// Check if forecast is still valid
	if wd.Meta.NextUpdate.Before(time.Now()) == true && assumeValid == false {
		os.Remove(cacheFile)
		return WeatherData{}, false
	}
	return wd, true
}

func fillWeatherDataCache(place string) (err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", yrURL(place), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "woozy, https://github.com/gummiboll/woozy")
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Failed to load forecast from yr.no")
	}

	defer resp.Body.Close()
	out, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

// LoadWeatherData loads xml from yr and returns a struct
func LoadWeatherData(place string, cc bool) (wd WeatherData, err error) {
	wd, valid := weatherDataCache(cc, false)
	if valid != true {
		err := fillWeatherDataCache(place)
		if err != nil {
			return wd, err
		}
		wd, valid := weatherDataCache(false, true)
		if valid != true {
			return wd, errors.New("Failed to load forecast")
		}
		return wd, nil
	}
	return wd, nil
}
