package main

import (
	"bufio"
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"text/template"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}()

var Build string

//go:embed templates/*.tmpl
var folder embed.FS

type DecoderType int

const (
	Base64Decoder DecoderType = iota
	AESDecoder
	XORDecoder
)

func (d DecoderType) String() string {
	return [...]string{"base64", "aes", "xor"}[d]
}

func StringToDecoderType(s string) DecoderType {
	switch s {
	case "base64":
		return Base64Decoder
	case "aes":
		return AESDecoder
	case "xor":
		return XORDecoder
	default:
		return Base64Decoder
	}
}

func main() {
	decoder := flag.String("d", "", "File containing the javascript decoder function (Default to embedded base64Decoder.tmpl)")
	tmpl := flag.String("t", "", "File containing the template for the output html (Default to embedded index.tmpl)")
	loot := flag.String("l", "", "File containing the contraband")
	label := flag.String("n", "", "Name of the contraband (Default to loot name)")
	output := flag.String("o", "index.html", "Output file")
	decoderType := flag.String("dt", "base64", "Decoder type (only base64 supported for now)")
	noBanner := flag.Bool("nb", false, "Disable the banner")
	ver := flag.Bool("v", false, "Print version")
	flag.Parse()
	if !*noBanner {
		Banner()
	}
	if *loot == "" {
		fmt.Println("No loot to smuggle")
		os.Exit(1)
	}
	if *label == "" {
		*label = *loot
	}
	if *ver {
		version()
		os.Exit(0)
	}

	decoderTypeValue := StringToDecoderType(*decoderType)
	if decoderTypeValue != Base64Decoder {
		fmt.Println("Only base64 decoder is supported for now")
		os.Exit(1)
	}
	contraband := Contraband{
		Path: *loot,
	}
	err := contraband.Read()
	if err != nil {
		panic(err)
	}
	err = contraband.Pack(*label, *decoder)
	if err != nil {
		panic(err)
	}
	err = contraband.Ship(*output, *tmpl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Happy smuggling! Your ship is at %s\n", *output)
}

type Contraband struct {
	Path    string
	Data    []byte
	Encoded string
}

type Inventory struct {
	Name string
	Loot string
}

type Ship struct {
	Contraband string
}

func (c *Contraband) Read() error {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		return err
	}
	c.Data = data
	return nil
}

func (c *Contraband) Base64() string {
	return base64.StdEncoding.EncodeToString(c.Data)
}

func (c *Contraband) Pack(label string, tmplFile string) error {
	inv := Inventory{
		Name: label,
		Loot: c.Base64(),
	}
	var decoderTemplate string
	if tmplFile == "" { // Use embedded template
		data, err := folder.ReadFile("templates/base64Decoder.tmpl")
		if err != nil {
			return err
		}
		decoderTemplate = string(data)
	} else {
		data, err := os.ReadFile(tmplFile)
		if err != nil {
			return err
		}
		decoderTemplate = string(data)
	}
	tmpl, err := template.New("HTML").Parse(decoderTemplate)
	if err != nil {
		return err
	}
	// Create a bytes.Buffer
	var buf strings.Builder

	// Write to the buffer through io.Writer interface
	writer := io.Writer(&buf)
	err = tmpl.Execute(writer, inv)
	if err != nil {
		return err
	}
	c.Encoded = buf.String()
	return nil
}

func (c *Contraband) Ship(path string, tmplFile string) error {
	ship := Ship{
		Contraband: c.Encoded,
	}
	var htmlTemplate string
	if tmplFile == "" { // Use embedded template
		data, err := folder.ReadFile("templates/index.tmpl")
		if err != nil {
			return err
		}
		htmlTemplate = string(data)
	} else {
		data, err := os.ReadFile(tmplFile)
		if err != nil {
			return err
		}
		htmlTemplate = string(data)
	}
	tmpl, err := template.New(tmplFile).Parse(htmlTemplate)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	err = tmpl.Execute(buf, ship)
	if err != nil {
		return err
	}
	buf.Flush()
	return nil
}

func Banner() {
	fmt.Println(`
	⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣠⠤⠶⠶⠾⠿⢗⠒⠦⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡤⠞⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠣⡀⢀⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣴⠾⠷⠶⢦⣤⣴⣶⣶⣶⠶⠖⠒⠒⠒⠒⠀⠀⠀⠀⠀⠀⠉⠉⠓⠒⠒⠤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡴⢋⣀⠄⠀⠈⠉⠉⠉⠁⠀⠀⠀⠀⠀⠤⠄⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠢⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣾⡿⠛⣁⣠⣤⣤⣤⣤⣄⣀⡀⠀⠠⣀⣀⣀⣀⣈⣿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡀⠀⠈⡄⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠟⢁⣴⣿⣟⣉⣉⠉⠉⠉⠉⠀⠀⠀⠀⠈⠛⠛⠛⠛⢛⣻⣿⣦⣀⣀⣀⣀⠀⠀⠀⠀⠀⠈⢄⠀⠹⡀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⠊⠀⠋⣉⣤⣾⠿⠿⠿⠶⣤⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⠛⠙⢧⡀⠉⠀⠀⠀⠀⠀⠀⠑⡄⢧⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⠃⣠⡴⠛⠉⣁⣀⣀⣀⣀⣀⢀⣙⣿⣿⣿⣿⣟⣋⣭⣍⣀⠀⢿⣿⣾⣿⣿⠦⡀⠀⠀⣄⠈⠲⢤⣀⣄⠈⠈⢧⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠃⣼⠋⠀⠐⣫⠿⠛⠛⢛⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣈⣿⣿⣿⣛⠳⣄⠀⠀⢉⠉⠲⢄⠈⠻⣷⡀⠘⣇⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠃⣼⠃⠀⠀⢀⠁⢀⣤⣾⠟⠋⠁⠀⣀⣤⣾⣿⠿⠿⠟⠛⠛⠛⠛⠛⠿⣿⠟⠻⣷⡘⣧⠰⣸⡄⠰⣄⠀⠀⠹⣿⡀⠸⡄⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡟⣼⣟⣠⠖⢠⣿⣠⡿⠋⠀⠀⠀⡰⠞⢻⡿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢷⣼⣿⣿⣿⣦⡈⢿⣦⣀⣿⣇⠀⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⣷⣿⡿⠃⠀⣸⢏⡿⠁⠀⠀⠀⠀⠀⠀⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⡏⢿⣿⣿⣿⣼⣿⣿⣿⣿⡀⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢀⡏⣿⣿⣿⣁⣠⡄⠀⣼⠃⠀⠀⠀⠀⠀⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⡇⡿⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⡞⢸⡟⣿⣿⣿⡿⢠⣾⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣇⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⡼⠀⢬⣤⣿⣿⣿⣿⢻⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⡇⣤⣴⣿⣿⣿⣿⡏⣾⡟⠀⠀⠐⠋⠉⠛⢶⣦⣤⣤⣤⣀⠀⠀⠀⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢇⠹⣿⣿⣿⣿⣿⡇⣿⠁⠀⠀⠀⠀⠶⠋⢩⣯⣿⠛⢧⡈⠀⠀⠀⠈⠀⣀⣀⣤⣶⣶⣾⣿⣿⣷⣤⡀⢀⣶⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⡎⢐⣋⣭⣿⡿⠸⣿⡏⠀⠀⠀⠀⠀⠀⠀⠢⠈⠀⠀⠒⠀⠀⠀⠀⠀⢠⣿⠟⣟⡛⠻⣝⡿⠻⢿⣿⣿⠿⠛⢻⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢻⠿⢿⣿⣿⠀⠀⠹⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⠘⠀⠀⠀⠀⠠⢀⣀⡠⠞⠉⠁⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣀⠴⠊⣡⣶⣿⣿⣿⣧⡀⢠⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⠛⡻⠀⠀⠀⠀⠀⠀
⢀⡤⠊⣁⣴⣿⣿⣿⣿⣿⡟⢡⣿⠏⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⠀⠀⠀⠘⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⡿⠿⡿⠁⢠⠃⠀⠀⠀⠀⠀⠀
⢋⣠⣾⣿⣿⣿⣿⣿⣿⠏⢠⣿⡟⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⡠⠊⠀⠀⠀⠀⠀⠀⠀⠙⢳⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⡟⢀⣀⢀⡴⠃⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⢸⣿⠃⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠞⠁⠀⢤⠄⠀⠀⠀⣀⡀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⣿⡟⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⢸⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠿⠟⠋⠙⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⠁⠀⢸⡇⠀⠀⠀⢄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⡄⠀⠀⠀⠀⠀⠀⠘⡄⠀⠀⠀⠀⠀⠀⣰⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠘⣧⠀⠀⠀⠈⣆⠀⠀⠀⠀⠠⠚⠓⠲⠶⠶⢦⣤⣀⣤⣀⡀⠀⠀⠀⠀⠈⠀⠀⠀⠀⢀⡔⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⡆⠀⠀⢻⡄⠀⠀⠀⠘⢆⠀⠀⠀⠀⠀⠀⠀⠄⡀⠀⠈⠉⠉⠉⠉⠋⡉⠓⢠⠀⠀⠀⢀⡴⡋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠈⢷⡀⠀⠀⠀⠈⢧⡀⠀⠀⠀⠀⠀⠀⠈⠉⠒⠒⠒⠒⠒⠉⠀⠀⠈⠀⠀⣠⣾⣆⠈⢦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⢻⣿⡀⠀⠀⠈⢷⡀⠀⠀⠀⠈⠻⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠀⠀⢀⣠⣾⣿⣝⢎⣳⣄⠑⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⡇⢸⣿⣷⡀⠀⠀⠀⢳⡀⠀⠀⠀⠀⠘⢷⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣴⡏⢹⡘⣿⣿⣿⣿⡏⠢⡈⠣⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣿⠏⣠⣾⣿⣿⣷⠀⢠⠀⠀⠹⣄⠀⠀⠀⠀⠈⢿⠒⠤⢤⣤⣤⣦⣤⣤⣤⣤⣴⣿⣿⡇⠀⣇⣿⣿⣿⣿⣿⡄⠙⢆⠙⢄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⡿⠛⠁⠚⠉⣿⣿⣿⣿⣧⡀⠡⠀⠀⠙⣦⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⠿⢿⣿⣿⣿⣿⣿⡇⢰⣿⢸⡿⠿⠿⠿⠟⠀⠀⠳⡌⢦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⡇⠈⠀⠀⠀⠀⠀⣿⠁⢹⣿⣿⣷⡄⠱⡀⠀⠈⢳⡀⠀⠀⠀⠀⠀⠀⠀⢀⣀⡠⢾⣿⣿⣿⡿⠋⢠⡿⢿⣾⡇⠀⠀⠀⠀⠀⠀⠀⠈⢦⠻⡕⠢⢄⡀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⣿⠀⠀⣿⣿⣿⣿⣦⠈⠀⠀⠀⠙⣦⠀⠀⠀⠀⠀⠉⠁⠀⠀⣠⣾⡿⠋⠀⢠⡟⠀⠀⢻⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⢳⠙⣆⠀⠈⢢⡀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⣿⠀⠀⢻⣿⣿⣿⣿⣷⣄⠀⠀⠀⠈⠳⣄⠀⠀⠀⠀⠀⣠⣾⣿⡟⠀⢀⣴⣿⠁⠀⠀⠈⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⣠⠏⠀⠀⠀⢳⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠙⠀⠀⠸⣿⣿⣿⣿⣿⣿⣿⣦⣄⡀⢀⡼⠀⠀⠀⣠⠟⠁⠀⣽⠃⢀⣾⣿⡇⠀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⢀⡠⠚⠁⠀⠀⠀⠀⠀⢧⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠀⠀⠀⠀⠀⠀⢠⡿⠀⢸⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡴⠋⠀⠀⠀⠀⠀⠀⢠⡄⠈⡆⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⡼⠁⠀⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⢀⡴⠋⠀⠀⠀⠀⠀⠀⠀⠀⢸⡛⠀⢸⡀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⡼⠁⠀⠀⠹⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠈⠻⢄⡀⠀⠀⠀⠀⠀⠀⠀⠀⣟⠛⠆⠀⡇
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⡼⠁⠀⠀⠀⠀⠹⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠲⢤⡀⠀⠀⠀⠀⠀⡟⠷⡀⠀⢁
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⡼⠁⠀⠀⠀⠀⠀⠀⠘⢷⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠳⠄⠀⠀⠀⣿⠀⠈⡆⢸
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⡼⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⠀⠀⠀⠀⣿⠀⠸⠁⢸
	`)
	fmt.Println("------------------------------------------------------------")
	version()
	fmt.Println("------------------------------------------------------------")
}

func version() {
	fmt.Printf("Han build %s by Mortimus\n", Commit)
}
