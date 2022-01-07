package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"reflect"
	//"unicode/utf8"

	"github.com/rabbitmq/amqp091-go"
)

var MsgsJson []string
var iCntMsg int = 0

var (
	uri         = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	insecureTLS = flag.Bool("insecure-tls", false, "Insecure TLS mode: don't check certificates")
	queue       = flag.String("queue", "", "AMQP queue name")
	ack         = flag.Bool("ack", false, "Acknowledge messages")
	maxMessages = flag.Uint("max-messages", 1000, "Maximum number of messages to dump or 0 for unlimited")
	outputDir   = flag.String("output-dir", ".", "Directory in which to save the dumped messages")
	full        = flag.Bool("full", false, "Dump the message, its properties and headers")
	verbose     = flag.Bool("verbose", false, "Print progress")
)

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unused command line arguments detected.\n")
		flag.Usage()
		os.Exit(2)
	}
	err := dumpMessagesFromQueue(*uri, *queue, *maxMessages, *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// Abre la conexión con RabbitMQ con el protocolo AMQP
// ---------------------------------------------------
func dial(amqpURI string) (*amqp091.Connection, error) {
	verboseLog(fmt.Sprintf("Dialing %q", amqpURI))
	if *insecureTLS && strings.HasPrefix(amqpURI, "amqps://") {
		tlsConfig := new(tls.Config)
		tlsConfig.InsecureSkipVerify = true
		conn, err := amqp091.DialTLS(amqpURI, tlsConfig)
		return conn, err
	}
	conn, err := amqp091.Dial(amqpURI)
	return conn, err
}

func dumpMessagesFromQueue(amqpURI string, queueName string, maxMessages uint, outputDir string) error {
	if queueName == "" {
		return fmt.Errorf("Must supply queue name")
	}

	// Inicia la conexión con RabbitMQ
	conn, err := dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	defer func() {
		conn.Close()
		verboseLog("AMQP connection closed")
	}()

	// Abre el canal con RabbitMQ
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	verboseLog(fmt.Sprintf("Pulling messages from queue %q", queueName))
	
	// Lee cada uno de los mensajes de la cola especificada en el paramero -queue
	for messagesReceived := uint(0); maxMessages == 0 || messagesReceived < maxMessages; messagesReceived++ {

		// Obtiene un mensaje de RabbitMQ
		msg, ok, err := channel.Get(queueName,
			*ack, // autoAck
		)
		if err != nil {
			return fmt.Errorf("Queue get: %s", err)
		}

		if !ok {
			verboseLog("No more messages in queue")
			break
		}

		/*
		// Guarda el mensaje (binario) en un archivo
		err = saveMessageToFile(msg.Body, outputDir, messagesReceived)
		if err != nil {
			return fmt.Errorf("Save message: %s", err)
		}
		*/

		// Si se indico la opcion -full, guarda el mensaje (JSON) en un archivo.
		if *full {
			//err = savePropsAndHeadersToFile(msg, outputDir, messagesReceived)
			err = ConvierteMsg2Json(msg, outputDir, messagesReceived)
			if err != nil {
				return fmt.Errorf("Save props and headers: %s", err)
			}
			
			if iCntMsg > 0 {
				GrabaMsgs()
			}
		}
	}

	return nil
}


func GrabaMsgs(){

	var nomArchJson = fmt.Sprintf("%s.json", queue)
	
	fmt.Println("nomArchJson: %s", nomArchJson)
	fmt.Println(reflect.TypeOf(nomArchJson))
	
fmt.Println(MsgsJson)
//s := []string(MsgsJson)
	//fmt.Println(utf8.RuneCountInString(string(s)))
	
	//byte[] barrImg = (byte[]) MsgsJson;
	//filePath := generateFilePath("./", nomArchJson)
/*
hello := []byte(MsgsJson)

	err := ioutil.WriteFile(nomArchJson, hello, 0644)
	if err != nil {
		return err
	}
*/	
	fmt.Println("Número de mensajes: ", iCntMsg)
//	fmt.Println("Archivo de mensajes: ", nomArchJson)	
}


// Convierte cada uno de los mensajes a formato JSON y lo gurda en un archivo.
func ConvierteMsg2Json(msg amqp091.Delivery, outputDir string, counter uint) error {
	extras := make(map[string]interface{})
	extras["properties"] = getProperties(msg)
	extras["headers"] = msg.Headers

	// Convierte el mensaje a formato JSON
	data, err := json.MarshalIndent(extras, "", "  ")
	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Agrega el mensaje en formato JSON a la variable MsgsJson
	// --------------------------------------------------------
	
	// Agrega el separador (coma) en la MsgsJson cuando el siguiente mensaje no sea el primero.
	if iCntMsg > 0 {
		MsgsJson = append(MsgsJson, ",")
	}

	// Incrementa el contador de mensajes y agrega el mensaje a la variable MsgsJson
	iCntMsg++
	MsgsJson = append(MsgsJson,  string(data))

	// --------------------------------------------------------

	return nil
}

// Graba cada mensaje binario en un archivo.
func saveMessageToFile(body []byte, outputDir string, counter uint) error {
	filePath := generateFilePath(outputDir, counter)
	err := ioutil.WriteFile(filePath, body, 0644)
	if err != nil {
		return err
	}

	// Muestra en pantalla el nombre del archivo con el mensaje binario.
	fmt.Println(filePath)

	return nil
}

func getProperties(msg amqp091.Delivery) map[string]interface{} {
	props := map[string]interface{}{
		"app_id":           msg.AppId,
		"content_encoding": msg.ContentEncoding,
		"content_type":     msg.ContentType,
		"correlation_id":   msg.CorrelationId,
		"delivery_mode":    msg.DeliveryMode,
		"expiration":       msg.Expiration,
		"message_id":       msg.MessageId,
		"priority":         msg.Priority,
		"reply_to":         msg.ReplyTo,
		"type":             msg.Type,
		"user_id":          msg.UserId,
		"exchange":         msg.Exchange,
		"routing_key":      msg.RoutingKey,
	}

	if !msg.Timestamp.IsZero() {
		props["timestamp"] = msg.Timestamp.String()
	}

	for k, v := range props {
		if v == "" {
			delete(props, k)
		}
	}

	return props
}

// Convierte cada uno de los mensajes a formato JSON y lo gurda en un archivo.
func savePropsAndHeadersToFile(msg amqp091.Delivery, outputDir string, counter uint) error {
	extras := make(map[string]interface{})
	extras["properties"] = getProperties(msg)
	extras["headers"] = msg.Headers

	// Convierte el mensaje a formato JSON
	data, err := json.MarshalIndent(extras, "", "  ")
	if err != nil {
		return err
	}

	// Genera un nuevo archivo y guarda en este el mensaje (JSON).
	filePath := generateFilePath(outputDir, counter) + "-headers+properties.json"
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	
	
	// Muestra en pantalla el nombre del archivo con el mensaje JSON.
	fmt.Println(filePath)

	return nil
}


func generateFilePath(outputDir string, counter uint) string {
	return path.Join(outputDir, fmt.Sprintf("msg-%04d", counter))
}

func verboseLog(msg string) {
	if *verbose {
		fmt.Println("*", msg)
	}
}



