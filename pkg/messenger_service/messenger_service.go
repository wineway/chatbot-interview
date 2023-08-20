package messenger_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/wineway/chatbot/pkg/message_service"
)

const SendMessageURL = "https://graph.facebook.com/v17.0/me/messages"

type MessengerService struct {
	token          string
	accessToken    string
	messageService message_service.MessageService
	mux            *http.ServeMux
}

type Message struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type User struct {
	ID int64 `json:"id,string"`
}

type MessageInfo struct {
	Sender    User            `json:"sender"`
	Recipient User            `json:"recipient"`
	Timestamp int64           `json:"timestamp"`
	Message   *MessageContent `json:"message"`
}

type Entry struct {
	ID        int64         `json:"id,string"`
	Time      int64         `json:"time"`
	Messaging []MessageInfo `json:"messaging"`
}

type MessageContent struct {
	Mid  string `json:"mid"`
	Text string `json:"text"`
}

type MessagingType string

const MessagingTypeResponse = "RESPONSE"

type TextMessage struct {
	Text string `json:"text"`
}

type MessageResponse struct {
	Recipient     User          `json:"recipient"`
	MessagingType MessagingType `json:"messaging_type"`
	Message       TextMessage   `json:"message"`
}

func NewMessageResponse(recipient int64, text string) MessageResponse {
	return MessageResponse{
		Recipient: User{
			ID: recipient,
		},
		MessagingType: MessagingTypeResponse,
		Message: TextMessage{
			Text: text,
		},
	}
}

func NewMessengerService(token string, accessToken string) MessengerService {
	mux := http.NewServeMux()
	s := MessengerService{
		token:          token,
		accessToken:    accessToken,
		messageService: &message_service.SampleMessageService{},
		mux:            mux,
	}
	s.mux.HandleFunc("/", s.handle)
	return s
}

func (s *MessengerService) Handler() http.Handler {
	return s.mux
}

func (s *MessengerService) handle(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == "GET" {
		s.Verify(r.FormValue("hub.verify_token"), r.FormValue("hub.challenge"), w)
	}

	var msg Message

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, &msg)

	if err != nil {
		// TODO
		return
	}

	s.HandleMessages(&msg)

	respond(w, http.StatusAccepted)

}

func (s *MessengerService) HandleMessage(msg *MessageInfo) error {
	rawResp, err := s.messageService.HandleEvent(&message_service.SimpleRequest{
		Sender:  message_service.UserInfo{},
		Content: msg.Message.Text,
	})

	if err != nil {
		return err
	}

	if rawResp == nil {
		// maybe do not need response
		return nil
	}

	response := NewMessageResponse(msg.Sender.ID, rawResp.Data)

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	log.Printf("data: %s", string(data))

	req, err := http.NewRequest("POST", SendMessageURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = "access_token=" + s.accessToken

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}

	log.Printf("resp code: %d\n", resp.StatusCode)

	return checkFacebookError(req.Body)
}

func (s *MessengerService) HandleMessages(msgs *Message) {
	for _, entry := range msgs.Entry {
		for _, msg := range entry.Messaging {
			err := s.HandleMessage(&msg)
			if err != nil {
				log.Println(fmt.Sprintf("handle msg: %v encounter an error: ", msg), err)
			}
		}
	}

}

func (s *MessengerService) Verify(token string, challenge string, w http.ResponseWriter) {
	var body []byte
	if token == s.token {
		body = []byte(challenge)
	} else {
		body = []byte("Incorrect verify token.")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func respond(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code": %d, "status": "%s"}`, code, http.StatusText(code))
}

type QueryResponse struct {
	Error  *QueryError `json:"error,omitempty"`
	Result string      `json:"result,omitempty"`
}

type QueryError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

func checkFacebookError(r io.Reader) error {
	var err error

	body, _ := io.ReadAll(r)
	log.Println(fmt.Sprintf("send response: %s", string(body)))
	r = io.NopCloser(bytes.NewBuffer(body))

	qr := QueryResponse{}
	err = json.NewDecoder(r).Decode(&qr)
	if err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}
	if qr.Error != nil {
		return fmt.Errorf("facebook error: %w", qr.Error)
	}

	return nil
}
