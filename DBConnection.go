package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataPoint struct {
	RemoteAdd string
	Points    string
}

func readPackets(s socketio.Conn) {

	p := make([]byte, 1024)

	addr := net.UDPAddr{
		Port: 41181,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	prevData := ""
	db := getSqlConnection()
	collection := getMongoDBConnection()
	for {
		n, remoteaddr, err := ser.ReadFromUDP(p)
		msg := strings.Split(string(p[:n]), ",")
		//fmt.Printf("Prev:  %s %s \n", prevData, msg[1])
		if prevData != msg[1] {
			fmt.Printf("Read a message from %v %s \n", remoteaddr, p[:n])
			prevData = msg[1]
			insertValueToDB(collection, msg[1])
			insertIntoMySQL(db, msg[1])
			s.Emit("field", string(p[:n]))

		}
		if err != nil {
			fmt.Printf("Client error  found :  %v\n", err)

			return
		}

	}
	defer db.Close()
}

func insertValueToDB(collection *mongo.Collection, res string) {

	record := DataPoint{time.Now().Format("2006.01.02 15:04:05"), res}

	insertResult, err := collection.InsertOne(context.TODO(), record)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

}

func insertIntoMySQL(db *sql.DB, res string) {

	fmt.Println(time.Now().Format("2006.01.02 15:04:05"))
	insert, err := db.Query("INSERT INTO DataPoint (recordDate,point)VALUES ('" + time.Now().Format("2006-01-02 15:04:05") + "', '" + res + "' )")

	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Record inserted into Mysql ")
	defer insert.Close()

}

func getSqlConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:icl=12321@/test")

	if err != nil {
		log.Fatal(err)
		return nil
	}
	fmt.Println("Connected to MySQL!")
	return db
}

func getMongoDBConnection() *mongo.Collection {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("GoDB").Collection("datapoint")
	return collection
}
func main() {

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")

		fmt.Println("connected:", s.ID())
		readPackets(s)
		s.Close()

		return nil
	})
	server.OnEvent("/msg", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)

		fmt.Println("connected:", s.ID())

		readPackets(s)

	})
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		fmt.Println("connected:", s.ID())
		readPackets(s)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset1")))
	log.Println("Serving at localhost:4000...")
	log.Fatal(http.ListenAndServe(":4000", nil))
}
