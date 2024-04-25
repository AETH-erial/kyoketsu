package kyoketsu

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
)

type ScanRequest struct {
	IpAddress string `json:"ip_address"`
}

// Holding all static web server resources
//
//go:embed html/bootstrap-5.0.2-dist/js/* html/bootstrap-5.0.2-dist/css/* html/* html/templates/*
var content embed.FS

/*
Run a new webserver

	:param port: port number to run the webserver on
*/
func RunHttpServer(port int, dbhook TopologyDatabaseIO, portmap []int) {
	assets := &AssetHandler{Root: content, RelPath: "static", EmbedRoot: "html"}
	tmpl, err := template.ParseFS(content, "html/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	iptable, err := template.ParseFS(content, "html/templates/ip_table.html")
	if err != nil {
		log.Fatal(err)
	}
	htmlHndl := &HtmlHandler{Home: tmpl, DbHook: dbhook}
	execHndl := &ExecutionHandler{DbHook: dbhook, PortMap: portmap, TableEntry: iptable}
	http.Handle("/static/", assets)
	http.Handle("/home", htmlHndl)
	http.Handle("/refresh", execHndl)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

}

type ExecutionHandler struct {
	DbHook     TopologyDatabaseIO
	TableEntry *template.Template
	PortMap    []int
}

func (e *ExecutionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input ScanRequest
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Fprintf(w, "There was an error processing your request: %s", err)
		return
	}
	err = json.Unmarshal(b, &input)
	if err != nil {
		fmt.Fprintf(w, "There was an error processing your request: %s", err)
		return

	}
	subnetMap, err := GetNetworkAddresses(input.IpAddress)
	if err != nil {
		fmt.Fprintf(w, "There was an error processing your request: %s", err)
	}
	scanned := make(chan Host)
	go func() {
		for x := range scanned {
			if len(x.ListeningPorts) > 0 {
				e.TableEntry.Execute(w, x)

				fmt.Print(" |-|-|-| :::: HOST FOUND :::: |-|-|-|\n==================||==================\n")
				fmt.Printf("IPv4 Address: %s\nListening Ports: %v\n=====================================\n", x.IpAddress, x.ListeningPorts)
				host, err := e.DbHook.GetByIP(x.IpAddress)
				if err != nil {
					if err != ErrNotExists {
						log.Fatal(err, " Couldnt access the database. Fatal error.\n")
					}
					_, err = e.DbHook.Create(x)
					if err != nil {
						log.Fatal(err, " Fatal error trying to read the database.\n")
					}
					continue
				}
				_, err = e.DbHook.Update(host.Id, x)
				if err != nil {
					log.Fatal(err, " fatal error when updating a record.\n")
				}

			}
		}
	}()
	NetSweep(subnetMap.Ipv4s, RetrieveScanDirectives(), scanned)
}

// handlers //

type HtmlHandler struct {
	Home   *template.Template // pointer to the HTML homepage
	DbHook TopologyDatabaseIO
}

/*
Handler function for HTML serving

	:param w: http.ResponseWriter interface for sending data back
	:param r: pointer to the http.Request coming in
*/
func (h *HtmlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.DbHook.All()
	if err != nil {
		fmt.Fprintf(w, "You have made it to the kyoketsu web server!\nThere was an error getting the db table, though.\n%s", err)
	}
	h.Home.Execute(w, data)
}

type AssetHandler struct {
	Root      embed.FS // Should be able to use anything that implements the fs.FS interface for serving asset files
	EmbedRoot string   // This is the root of the embeded file system
	RelPath   string   // The path that will be used for the handler, relative to the root of the webserver (/static, /assets, etc)
}

/*
Handler function to serve out asset files (HTMX, bootstrap, pngs etc)

	:param w: http.ResponseWriter interface for sending data back to the caller
	:param r: pointer to an http.Request
*/
func (a *AssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var uripath string     // the path from the request
	var pathSp []string    // the path from the request split, so that we can point the request path to the embedded fs
	var assetPath []string // the cleaned path for the requested asset
	var fname string       // filename of the requested asset
	var ctype string
	var b []byte
	var err error

	uripath = strings.TrimPrefix(r.URL.Path, a.RelPath)
	uripath = strings.Trim(uripath, "/")
	pathSp = strings.Split(uripath, "/")
	fname = pathSp[len(pathSp)-1]
	assetPath = append(assetPath, a.EmbedRoot)
	for i := 1; i < len(pathSp); i++ {
		assetPath = append(assetPath, pathSp[i])
	}
	b, err = a.Root.ReadFile(path.Join(assetPath...))
	if err != nil {
		fmt.Fprintf(w, "Error occured: %s. path split: '%s'\nAsset Path: %s", err, pathSp, assetPath)
	}
	switch {
	case strings.Contains(fname, "css"):
		ctype = "text/css"
	case strings.Contains(fname, "js"):
		ctype = "text/javascript"
	case strings.Contains(fname, "html"):
		ctype = "text/html"
	case strings.Contains(fname, "json"):
		ctype = "application/json"
	case strings.Contains(fname, "png"):
		ctype = "image/png"
	default:
		ctype = "text"
	}
	w.Header().Add("Content-Type", ctype)
	fmt.Fprint(w, string(b))

}
