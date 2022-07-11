package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
)

type Table struct {
	XMLName xml.Name `xml:"table"`
	Text    string   `xml:",chardata"`
	Name    string   `xml:"name,attr"`
	Tag     []struct {
		Name     string `xml:"name,attr"`
		Type     string `xml:"type,attr"`
		Writable string `xml:"writable,attr"`
		Desc     []struct {
			Text string `xml:",chardata"`
			Lang string `xml:"lang,attr"`
		} `xml:"desc"`
	} `xml:"tag"`
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

var regexsplit = regexp.MustCompile("</table>")
var regexsplit1 = regexp.MustCompile("<taginfo>")
var channel chan []byte

func ScanByTable(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if loc, loc1 := regexsplit.FindIndex(data), regexsplit1.FindIndex(data); loc != nil && loc1 != nil {
		return loc[1] + 1, data[loc1[1]+1 : loc[1]], nil
	}
	if loc := regexsplit.FindIndex(data); loc != nil {
		return loc[1] + 1, data[0:loc[1]], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func generateOutput(channel chan []byte) {
	cmd := exec.Command("exiftool", "-listx")
	stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	scanner.Split(ScanByTable)
	xmlinput := Table{}
	jsonOutput := JSONOutput{}

	go func() {
		for scanner.Scan() {
			m := scanner.Text()
			xml.Unmarshal([]byte(m), &xmlinput)

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
			}
			b, _ := json.Marshal(jsonOutput)
			channel <- b
		}
		close(channel)
	}()

}

func myhandler(w http.ResponseWriter, r *http.Request) {

	go generateOutput(channel)
	for n := range channel {
		b := string(n)
		fmt.Fprintf(w, b)
		w.(http.Flusher).Flush()
	}
}
func main() {
	channel = make(chan []byte)
	http.HandleFunc("/", myhandler)
	http.ListenAndServe(":8080", nil)
}
