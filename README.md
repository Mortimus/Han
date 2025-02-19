# Han
a simple HTML smuggle page generator

## Usage
```bash
Han -d base64Decoder.tmpl -l malicious.exe -n notmalicious.exe -o index.html -t index.tmpl
```

## Options
```
  -d string
        File containing the javascript decoder function (default "base64Decoder.tmpl")
  -l string
        File containing the contraband (default "loot.txt")
  -n string
        Name of the contraband (default "loot")
  -o string
        Output file (default "index.html")
  -t string
        File containing the template for the output html (default "template.tmpl")
```