/*
Package server provides functionality for two kind of servers
 - TCP thermomatic  protocol
 - HTTP json healthcheck endpoints

The TCP thermomatic server uses goroutines to handle each connected
device.  3 channels are used to communicate the client data to the server core

	commands chan common.Command
	    used to send data send by the client to the server core.
	    currently the server implements only one command (READING)
	Logouts  chan *client.Client
		used to send clients with closed or timeout connections. The reciever
		should remove the record from the connected clients map
	Logins   chan *client.Client
		after login message is recived and parsed as imei this channel is used
		to send a newly created(and) client so the receiver can store
		its reference in the connected clients map

These HTTP are the implemented json endpoints

  - `GET /stats`: returns a JSON document which contains runtime statistical
     information about the server (i.e. number of goroutines, bytes read per second, etc.).
  - `GET /readings/:imei:` if the device is online returns a JSON representation of
     the last reading the device has sent (timestamped)
  - `GET /status/:imei:` reports whether the device is online or not.
*/
package server
