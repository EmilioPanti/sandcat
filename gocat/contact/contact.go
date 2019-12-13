package contact 

//Contact defines required functions for communicating with the server
type Contact interface {
	Ping(server string) bool
	GetInstructions(profile map[string]interface{}) map[string]interface{}
	DropPayloads(payload string, server string, payload_localname_map map[string]string) []string
	RunInstruction(command map[string]interface{}, profile map[string]interface{}, payloads []string, payload_localname_map map[string]string)
	C2RequirementsMet(criteria interface{}) bool
}

//CommunicationChannels contains the contact implementations
var CommunicationChannels = map[string]Contact{}