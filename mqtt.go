package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

// newTLSConfig constructs a TLS configuration for secure MQTT connections.
// It loads the bundled Victron root certificate and uses it to validate
// certificate chains when connecting to a hostname. When the broker is
// configured as a raw IP address, hostname verification is disabled because
// many Victron MQTT certificates do not include IP SANs.
func newTLSConfig(host string) *tls.Config {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

	cfg := &tls.Config{
		RootCAs: roots,
	}

	if net.ParseIP(host) == nil {
		cfg.ServerName = host
		cfg.InsecureSkipVerify = false
	} else {
		cfg.InsecureSkipVerify = true //nolint:gosec
	}

	return cfg
}

// mqttConnectionConfig stores configuration values required to connect to the
// Victron MQTT broker, including host, port, TLS, and credentials.
type mqttConnectionConfig struct {
	host     string
	port     int
	secure   bool
	username string
	password string
}

// connectWait blocks until the provided MQTT client finishes the connect attempt
// or returns an error. It centralizes connect timeout handling.
func connectWait(client mqtt.Client) error {
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}

	err := token.Error()
	if err != nil {
		return fmt.Errorf("failed to connect to mqtt: %w", err)
	}

	return nil
}

// connect creates and synchronously connects an MQTT client for publishing.
func connect(clientID string, config mqttConnectionConfig) (mqtt.Client, error) {
	client := mqtt.NewClient(createClientOptions(clientID, config, nil))

	return client, connectWait(client)
}

// listen creates an MQTT client that subscribes to topics after every successful
// connection event. The topic argument is passed directly to Subscribe.
func listen(clientID string, config mqttConnectionConfig, topic string) error {
	log.WithFields(log.Fields{
		"host": config.host,
		"port": config.port,
	}).Debug("connecting to mqtt")

	onConnect := func(client mqtt.Client) {
		log.Info("mqtt connected, subscribing to topics...")
		// We need to subscribe after each connection
		// since mqtt does not maintain subscriptions across reconnects
		client.Subscribe(topic, 0, mqttSubscriptionHandler)
	}

	client := mqtt.NewClient(createClientOptions(clientID, config, onConnect))

	return connectWait(client)
}

// newConnectionLostHandler returns a handler that is invoked when the MQTT
// client loses its connection. It updates Prometheus metrics and logs the
// failure so the exporter can be monitored.
func newConnectionLostHandler(clientID string) mqtt.ConnectionLostHandler {
	return func(c mqtt.Client, e error) {
		log.WithFields(log.Fields{
			"client_id": clientID,
		}).WithError(e).Error("mqtt connection lost")
		connectionStatus.WithLabelValues(clientID).Set(0)
		connectionStatusSinceTimeSeconds.WithLabelValues(clientID).Set(float64(time.Now().Unix()))
	}
}

// newConnectionHandler returns a handler invoked once the MQTT client reconnects.
// It updates Prometheus metrics and optionally invokes a wrapped callback to
// subscribe to topics or perform other initialization.
func newConnectionHandler(clientID string, wrapped mqtt.OnConnectHandler) mqtt.OnConnectHandler {
	return func(c mqtt.Client) {
		log.WithField("client_id", clientID).Info("mqtt connected")
		connectionStatus.WithLabelValues(clientID).Set(1)
		connectionStatusSinceTimeSeconds.WithLabelValues(clientID).Set(float64(time.Now().Unix()))

		if wrapped != nil {
			wrapped(c)
		}
	}
}

// createClientOptions constructs MQTT client options for the Eclipse Paho
// library, enabling auto-reconnect, connection monitoring, and optional TLS.
func createClientOptions(clientID string, config mqttConnectionConfig, onConnectionHandler mqtt.OnConnectHandler) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetWriteTimeout(30 * time.Second)
	opts.SetOrderMatters(false)
	opts.SetConnectionLostHandler(newConnectionLostHandler(clientID))
	opts.SetOnConnectHandler(newConnectionHandler(clientID, onConnectionHandler))

	if config.secure {
		opts.AddBroker(fmt.Sprintf("ssl://%s:%d", config.host, config.port))
		opts.SetTLSConfig(newTLSConfig(config.host))
	} else {
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.host, config.port))
	}

	if config.username != "" {
		opts.SetUsername(config.username)
	}

	if config.password != "" {
		opts.SetPassword(config.password)
	}

	opts.SetClientID(clientID)
	opts.SetCleanSession(true)

	return opts
}

// victronValue mirrors the JSON value payload used in most Victron MQTT messages.
// The optional pointer allows the exporter to distinguish between null and zero.
type victronValue struct {
	Value *float64 `json:"value"`
}

// victronStringValue mirrors the payload used for string-based Victron MQTT values
// such as the system serial number topic.
type victronStringValue struct {
	Value *string `json:"value"`
}

// mqttSubscriptionHandler parses messages arriving on subscribed Victron MQTT
// topics, maps them to Prometheus metrics, and ignores unsupported topics.
func mqttSubscriptionHandler(client mqtt.Client, msg mqtt.Message) {
	subscriptionsUpdatesTotal.Inc()

	topic := msg.Topic()
	topicParts := strings.Split(topic, "/")
	if len(topicParts) < 5 {
		subscriptionsUpdatesIgnoredTotal.Inc()

		return
	}
	topicInfoParts := topicParts[4:]

	componentType := topicParts[2]
	componentID := topicParts[3]

	topicString := strings.Join(topicInfoParts, "/")

	if (topicString == "Serial") && (systemSerialID == "") {
		var v victronStringValue

		err := json.Unmarshal(msg.Payload(), &v)
		if err != nil {
			log.Warn("failed to unmarshal victron mqtt payload: ", err)
			subscriptionsUpdatesIgnoredTotal.Inc()

			return
		}

		if v.Value != nil {
			systemSerialID = *v.Value
		}

		return
	}

	o, ok := suffixTopicMap[topicString]
	if !ok {
		log.Debug("mqtt topic '", topicString, "' not mapped to a metric, ignoring: ", topic)
		subscriptionsUpdatesIgnoredTotal.Inc()

		return
	}

	var v victronValue

	err := json.Unmarshal(msg.Payload(), &v)
	if err != nil {
		log.Warn("failed to unmarshal victron mqtt payload: ", err)
		subscriptionsUpdatesIgnoredTotal.Inc()

		return
	}

	if v.Value == nil {
		o(componentType, componentID, math.NaN())
	} else {
		o(componentType, componentID, *v.Value)
	}
}
