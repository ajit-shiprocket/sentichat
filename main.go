package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	
	_ "github.com/mattn/go-sqlite3"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waEvents "go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Chat struct {
	JID             string    `json:"jid"`
	Name            string    `json:"name"`
	LastMessageTime time.Time `json:"last_message_time"`
}

type Message struct {
	ID        string    `json:"id"`
	ChatJID   string    `json:"chat_jid"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsFromMe  bool      `json:"is_from_me"`
}

type MessageStore struct {
	db *sql.DB
}

var client *whatsmeow.Client
var msgStore *MessageStore
var qrCode string
var qrMutex sync.Mutex

// func NewMessageStore() (*MessageStore, error) {
// 	db, err := sql.Open("sqlite3", "file:store/messages.db?_foreign_keys=on")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &MessageStore{db: db}, nil
// }

func NewMessageStore() (*MessageStore, error) {
	// if err := os.MkdirAll("store", 0755); err != nil {
	// 	return nil, fmt.Errorf("failed to create store directory: %v", err)
	// }
	db, err := sql.Open("sqlite3", "file:store/messages.db?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open message database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS chats (
		  jid TEXT PRIMARY KEY,
		  name TEXT,
		  last_message_time TEXT
		);
	`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT,
			sender TEXT,
			content TEXT,
			timestamp TIMESTAMP,
			is_from_me BOOLEAN,
			media_type TEXT,
			filename TEXT
		);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}
	return &MessageStore{db: db}, nil
}

func (store *MessageStore) SaveIncomingMessage(msg Message) error {
	_, err := store.db.Exec(
		"INSERT INTO messages (id, chat_jid, sender, content, timestamp, is_from_me) VALUES (?, ?, ?, ?, ?, ?)",
		msg.ID, msg.ChatJID, msg.Sender, msg.Content, msg.Timestamp.Format(time.RFC3339), msg.IsFromMe,
	)
	return err
}

func (store *MessageStore) GetChats() ([]Chat, error) {
	rows, err := store.db.Query("SELECT jid, name, last_message_time FROM chats")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var c Chat
		var t string
		if err := rows.Scan(&c.JID, &c.Name, &t); err != nil {
			return nil, err
		}
		c.LastMessageTime, _ = time.Parse(time.RFC3339, t)
		chats = append(chats, c)
	}
	return chats, nil
}

func (store *MessageStore) GetMessages(chatJID string) ([]Message, error) {
	rows, err := store.db.Query("SELECT id, chat_jid, sender, content, timestamp, is_from_me FROM messages WHERE chat_jid = ?", chatJID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var t string
		if err := rows.Scan(&m.ID, &m.ChatJID, &m.Sender, &m.Content, &t, &m.IsFromMe); err != nil {
			return nil, err
		}
		m.Timestamp, _ = time.Parse(time.RFC3339, t)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func getChats(w http.ResponseWriter, r *http.Request) {
	chats, err := msgStore.GetChats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chats)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	jid := mux.Vars(r)["jid"]
	msgs, err := msgStore.GetMessages(jid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(msgs)
}

type SendMessageRequest struct {
	JID     string `json:"jid"`
	Message string `json:"message"`
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	jid := types.NewJID(req.JID, "s.whatsapp.net")
	msg := &waProto.Message{Conversation: proto.String(req.Message)}
	_, err := client.SendMessage(r.Context(), jid, msg)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "message sent"})
}

func getQRCode(w http.ResponseWriter, r *http.Request) {
	qrMutex.Lock()
	defer qrMutex.Unlock()
	if qrCode == "" {
		http.Error(w, "No QR code available or already authenticated", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"qr": qrCode})
}

func getQRCodeImage(w http.ResponseWriter, r *http.Request) {
	qrMutex.Lock()
	defer qrMutex.Unlock()
	if qrCode == "" {
		http.Error(w, "No QR code available or already authenticated", http.StatusNotFound)
		return
	}
	png, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR image", http.StatusInternalServerError)
		return
	}
	base64Image := base64.StdEncoding.EncodeToString(png)
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Image)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"qr_image": dataURL})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:store/whatsapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Fatal(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	client = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))

	// client.AddEventHandler(func(evt interface{}) {
	// 	switch v := evt.(type) {
	// 	case *waEvents.Message:
	// 		if v.Info.MessageSource.IsFromMe || v.Message.GetConversation() == "" {
	// 			return
	// 		}
	// 		incoming := Message{
	// 			ID:        v.Info.ID,
	// 			ChatJID:   v.Info.Chat.String(),
	// 			Sender:    v.Info.Sender.User,
	// 			Content:   v.Message.GetConversation(),
	// 			Timestamp: v.Info.Timestamp,
	// 			IsFromMe:  false,
	// 		}
	// 		_ = msgStore.SaveIncomingMessage(incoming)
	// 	}
	// })

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *waEvents.Message:
			msg := v.Message

			var content, mediaType, filename string

			switch {
			case msg.GetConversation() != "":
				content = msg.GetConversation()
				mediaType = "text"

			case msg.GetImageMessage() != nil:
				content = msg.GetImageMessage().GetCaption()
				mediaType = "image"

			case msg.GetVideoMessage() != nil:
				content = msg.GetVideoMessage().GetCaption()
				mediaType = "video"

			case msg.GetDocumentMessage() != nil:
				content = msg.GetDocumentMessage().GetTitle()
				mediaType = "document"
				filename = msg.GetDocumentMessage().GetFileName()

			default:
				content = "[Unsupported message type]"
				mediaType = "unknown"
			}

			sender := v.Info.Sender.User
			timestamp := v.Info.Timestamp
			isFromMe := v.Info.IsFromMe

			_, err := msgStore.db.Exec(`
				INSERT INTO messages (id, sender, content, timestamp, is_from_me, media_type, filename)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				v.Info.ID, sender, content, timestamp, isFromMe, mediaType, filename,
			)
			if err != nil {
				fmt.Println("DB insert error:", err)
			} else {
				fmt.Printf("Stored message from %s: %s\n", sender, content)
			}
		}
	})

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(ctx)
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					qrMutex.Lock()
					qrCode = evt.Code
					qrMutex.Unlock()
				} else if evt.Event == "success" || evt.Event == "timeout" || evt.Event == "error" {
					qrMutex.Lock()
					qrCode = ""
					qrMutex.Unlock()
				}
			}
		}()
		_ = client.Connect()
	} else {
		_ = client.Connect()
	}

	msgStore, err = NewMessageStore()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/chats", getChats).Methods("GET")
	r.HandleFunc("/chats/{jid}/messages", getMessages).Methods("GET")
	r.HandleFunc("/send", sendMessage).Methods("POST")
	r.HandleFunc("/qr/code", getQRCode).Methods("GET")
	r.HandleFunc("/qr", getQRCodeImage).Methods("GET")

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
