package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"os/exec"
)

var channel chan JSONOutput

type Taginfo struct {
	XMLName xml.Name `xml:"taginfo"`
	Text    string   `xml:",chardata"`
	Table   []struct {
		Text string `xml:",chardata"`
		Name string `xml:"name,attr"`
		G0   string `xml:"g0,attr"`
		G1   string `xml:"g1,attr"`
		G2   string `xml:"g2,attr"`
		Desc []struct {
			Text string `xml:",chardata"`
			Lang string `xml:"lang,attr"`
		} `xml:"desc"`
		Tag []struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,attr"`
			Name     string `xml:"name,attr"`
			Type     string `xml:"type,attr"`
			Writable string `xml:"writable,attr"`
			G2       string `xml:"g2,attr"`
			Count    string `xml:"count,attr"`
			Index    string `xml:"index,attr"`
			G1       string `xml:"g1,attr"`
			Desc     []struct {
				Text string `xml:",chardata"`
				Lang string `xml:"lang,attr"`
			} `xml:"desc"`
			Values []struct {
				Text  string `xml:",chardata"`
				Index string `xml:"index,attr"`
				Key   []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
					Val  []struct {
						Text string `xml:",chardata"`
						Lang string `xml:"lang,attr"`
					} `xml:"val"`
				} `xml:"key"`
			} `xml:"values"`
		} `xml:"tag"`
	} `xml:"table"`
}
type JSONOutput struct {
	Tag []Tags `json:"tags"`
}
type Tags struct {
	Writable string            `json:"writable"`
	Path     string            `json:"path"`
	Group    string            `json:"group"`
	Desc     map[string]string `json:"description"`
	Type     string            `json:"type"`
}

func getOutput(ctx context.Context) {

	cmd := exec.CommandContext(ctx, "exiftool", "-listx")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	go func() {
		m := Taginfo{}
		jsonOutput := JSONOutput{}

		dec := xml.NewDecoder(stdout)
		dec.Decode(&m)
		for _, xmlinput := range m.Table {
			for i := 0; i < len(xmlinput.Tag); i++ {
				languages := make(map[string]string)
				for j := 0; j < len(xmlinput.Tag[i].Desc); j++ {
					languages[xmlinput.Tag[i].Desc[j].Lang] = xmlinput.Tag[i].Desc[j].Text
				}
				x := Tags{}
				x.Path = (xmlinput.Name + ":" + xmlinput.Tag[i].Name)
				x.Group = xmlinput.Name
				x.Desc = languages
				x.Writable = xmlinput.Tag[i].Writable
				x.Type = xmlinput.Tag[i].Type
				jsonOutput.Tag = append(jsonOutput.Tag, x)
				channel <- jsonOutput
			}
		}
		close(channel)

	}()
	<-ctx.Done()
	//go io.Copy(os.Stdout, out)
}
func myHandler(w http.ResponseWriter, _ *http.Request) {
	log.Print("Client connected, returning json")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go getOutput(ctx)
	for n := range channel {
		err := json.NewEncoder(w).Encode(n)
		if err != nil {
			log.Print("client is gone, shutting down")
			<-ctx.Done()
			w.(http.Flusher).Flush()
		}
		w.(http.Flusher).Flush()
	}
}
func main() {
	channel = make(chan JSONOutput)
	http.HandleFunc("/", myHandler)
	http.ListenAndServe(":8080", nil)
}
