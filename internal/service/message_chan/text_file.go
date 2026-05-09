package message_chan

import "IMChat/internal/dto/request"

func handleText(req *request.ChatMessageRequest, deliver Deliver, norm NormalizePath) error {
	m := textEntity(req, norm)
	createMessage(&m)
	return routeTextOrFile(m, req, deliver)
}

func handleFile(req *request.ChatMessageRequest, deliver Deliver, norm NormalizePath) error {
	m := fileEntity(req, norm)
	createMessage(&m)
	return routeTextOrFile(m, req, deliver)
}
