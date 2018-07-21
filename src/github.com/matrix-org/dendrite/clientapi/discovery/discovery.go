package discovery

import (
	"github.com/matrix-org/dendrite/common/config"
	"time"
	"fmt"
	"bytes"
	"encoding/binary"
	"net"
	"github.com/sirupsen/logrus"
)

var advertise = false

func Start(config config.Dendrite) {
	advertise = config.Discovery.Enabled
	if advertise {
		go advertiseLoop(config)
	}
}

func Stop() {
	advertise = false
}

func advertiseLoop(config config.Dendrite) {
	advertiseAddr := "255.255.255.255"
	advertisePort := 8228

	advertiseStr := fmt.Sprintf("A;%s;P;%d;S;%s;N;%s;ents_rid_ss;%s;ents_rid_hgps;%s;ents_rid_hrmp;%s;ents_as_token;%s;ents_as_prefix;%s;ents_hs_domain;%s;",
		config.Discovery.ConnectIp,
		config.Discovery.ConnectPort,
		config.Discovery.ConnectScheme,
		config.Discovery.Name,
		config.Discovery.Ents.RoomIds.SuperSimon,
		config.Discovery.Ents.RoomIds.HovercraftGps,
		config.Discovery.Ents.RoomIds.HovercraftRamps,
		config.Discovery.Ents.Appservice.AdvertisedToken,
		config.Discovery.Ents.Appservice.Prefix,
		config.Matrix.ServerName)
	advertisePayload := []byte(advertiseStr)
	payloadSize := int32(len(advertisePayload))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, payloadSize)
	if err != nil {
		panic(err) // This is probably not great
	}

	packaged := make([]byte, 0)
	for _, b := range buf.Bytes() {
		packaged = append(packaged, b)
	}
	for _, b := range advertisePayload {
		packaged = append(packaged, b)
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", advertiseAddr, advertisePort))
	if err != nil {
		panic(err) // Not great
	}

	for advertise == true {
		c, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			logrus.Error(err)
		} else {
			c.Write(packaged)
			c.Close() // ignore error
		}
		time.Sleep(5 * time.Second)
	}
}
