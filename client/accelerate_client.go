package client

import (
	"fmt"
	"github.com/glvd/accipfs/core"
	"github.com/glvd/accipfs/general"
	"github.com/goextension/log"
	"strconv"
	"strings"
)

// ID ...
func ID(url string) (*core.NodeInfo, error) {
	reply := new(core.NodeInfo)
	if err := general.RPCPost(url, "Accelerate.ID", core.DummyEmpty(), reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// Ping ...
func Ping(info *core.NodeInfo) error {
	log.Debugw("ping info", "addr", info.RemoteAddr, "port", info.Port)
	pingAddr := strings.Join([]string{info.RemoteAddr, strconv.Itoa(info.Port)}, ":")
	url := fmt.Sprintf("http://%s/rpc", pingAddr)
	result := new(string)
	if err := general.RPCPost(url, "Accelerate.Ping", core.DummyEmpty(), result); err != nil {
		return err
	}
	if *result != "pong" {
		return fmt.Errorf("get wrong response data:%s", *result)
	}
	return nil
}

// Pins ...
func Pins(info *core.NodeInfo) ([]string, error) {
	log.Debugw("pin info", "addr", info.RemoteAddr, "port", info.Port)
	pingAddr := strings.Join([]string{info.RemoteAddr, strconv.Itoa(info.Port)}, ":")
	url := fmt.Sprintf("http://%s/rpc", pingAddr)
	result := new([]string)
	if err := general.RPCPost(url, "Accelerate.Pins", core.DummyEmpty(), result); err != nil {
		return nil, err
	}
	return *result, nil
}

// PinVideo ...
func PinVideo(url string, no string) error {
	log.Debugw("pin hash", "hash", no)
	b := new(bool)
	err := general.RPCPost(url, "Accelerate.PinVideo", &no, b)
	if err != nil {
		return err
	}
	if *b {
		fmt.Printf("pin (%s) success\n", no)
	}
	return nil
}

// Peers ...
func Peers(url string, info *core.NodeInfo) ([]*core.NodeInfo, error) {
	//pingAddr := strings.Join([]string{info.RemoteAddr, strconv.Itoa(info.Port)}, ":")
	//url := fmt.Sprintf("http://%s/rpc", pingAddr)
	result := new([]*core.NodeInfo)
	if err := general.RPCPost(url, "Accelerate.Peers", info, result); err != nil {
		return nil, err
	}
	if len(*result) == 0 {
		return nil, fmt.Errorf("no data response")
	}
	return *result, nil
}

// AddPeer ...
func AddPeer(url string, info *core.NodeInfo) error {
	status := new(bool)
	if err := general.RPCPost(url, "Accelerate.AddPeer", info, status); err != nil {
		log.Errorw("remote id error", "error", err.Error())
		return fmt.Errorf("remote id error: %w", err)
	}

	if !(*status) {
		return fmt.Errorf("connect failed:%s", url)
	}
	return nil
}
