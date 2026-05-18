package protocol

type MsgType int

// types of messages that can be sent to manager
const (
	MsgRequest MsgType = iota
	MsgGrant
	MsgRelease
	MsgHeartbeat
	MsgEvict
)

type Message struct {
	Id     string  `json:"id"`     //unique id for front end tracking
	NodeId string  `json:"NodeId"` //sender node's id
	Type   MsgType `json:"type"`   //mesaage type
	Token  int64   `json:"token"`  //token to filter stale packets
} //It tells Go’s JSON package (encoding/json) how this struct field should appear when converting between Go structs and JSON.