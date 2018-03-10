package util

import (
	"bufio"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
)

const (
	degRadConversion = math.Pi / 180
)

var templates *template.Template

//LoadTemplates initializes template
func LoadTemplates(htmlTemplates string) {
	templates = template.Must(template.ParseGlob(htmlTemplates))
}

//ExecuteTemplates passing data to html using comments; nil when nothing
func ExecuteTemplates(w http.ResponseWriter, htmlTemplates string, comments interface{}) {
	templates.ExecuteTemplate(w, htmlTemplates, comments)
}

func Scanner() *string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("enter the Google Maps API key")
	fmt.Println("-----------------------------")
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	return &text
}

func DegToRad(d float64) float64 { return d * degRadConversion }

func Odd(number int) bool { return number%2 != 0 }

func Round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5, f))
}

func RoundToF7(f float64) float64 {
	return math.Round(f*10000000) / 10000000
}
