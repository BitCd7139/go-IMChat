package message_chan

type Deliver func(userID string, wsPayload []byte, msgUUID string)

type NormalizePath func(path string) string
