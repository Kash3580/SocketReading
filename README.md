# Socket Reading

Reading packet from speficied socket port: This will read the packet from UDP port 4000 Continousely 


	func readPackets(s socketio.Conn) {
      p := make([]byte, 1024)

	addr := net.UDPAddr{
		Port: 4000,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	prevData := ""
	for {
		n, remoteaddr, err := ser.ReadFromUDP(p)
		msg := strings.Split(string(p[:n]), ",")
		fmt.Printf("Prev:  %s %s \n", prevData, msg[1])
		if prevData != msg[1] {
			fmt.Printf("Read a message from %v %s \n", remoteaddr, p[:n])
			prevData = msg[1]
			insertValueToDB(remoteaddr, msg[1]) // call function to insert the value in mongo db before emitting 
			s.Emit("field", string(p[:n]))

		}
		if err != nil {
			fmt.Printf("Client error  found :  %v\n", err)
			return
		}
    }

 

Creating new server to emit the data 

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
	 
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:4000...")
	log.Fatal(http.ListenAndServe(":4000", nil))
 
# Connecting MongoDB with GOLang

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("GoDB").Collection("datapoint")
	record := DataPoint{time.Now().Format("2006.01.02 15:04:05"), res}
	if err != nil {
		log.Fatal(err)
	}

	insertResult, err := collection.InsertOne(context.TODO(), record)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
