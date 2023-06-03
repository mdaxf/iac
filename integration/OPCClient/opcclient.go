package opcclient

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/open62541/open62541"
)

type OPCClient struct {
	Client    *open62541.Client
	Endpoint  string
	Namespace uint16
	Timeout   time.Duration
	Nodes     map[string]*open62541.Node
	iLog      logger.Log
}

func NewOpcClient(configurations OPCClient) *OPCClient {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OpcClient"}

	iLog.Debug(fmt.Sprintf(("Create OpcClient with configuration : %s"), logger.ConvertJson(configurations)))

	opcclient := &OPCClient{
		Endpoint:  configurations.Endpoint,
		Namespace: configurations.Namespace,
		Timeout:   configurations.Timeout,
		iLog:      iLog,
	}

	opcclient.Connect()

	return opcclient
}

func (c *OPCClient) Connect() error {
	c.iLog.Debug(fmt.Sprintf("Connectting to OPC Server: %s", c.Endpoint))
	client := open62541.NewClient(c.Endpoint)
	err := client.Connect()
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to connect to OPC Server: %v", err))
		return err
	}
	c.Client = client

	c.iLog.Debug(fmt.Sprintf("Connected to OPC Server: %s", c.Endpoint))

	return nil
}
func (c *OPCClient) ReadTag(tagID string) (interface{}, error) {
	c.iLog.Debug(fmt.Sprintf("Read Tag: %s NameSpace: %s", tagID, c.Namespace))
	nodeID, err := open62541.NewStringNodeID(c.Namespace, tagID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create the node id for Tag: %s", tagID))
		return nil, err
	}

	v, err := c.Client.ReadVariable(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to read Tag: %s", tagID))
		return nil, err
	}
	c.iLog.Debug(fmt.Sprintf("Read Tag: %s Value: %v", tagID, v.Value()))
	return v.Value.Value(), nil
}

func (c *OPCClient) WriteTag(tagID string, value interface{}) error {
	c.iLog.Debug(fmt.Sprintf("Write Tag: %s NameSpace: %s Value: %v", tagID, c.Namespace, value))
	nodeID, err := open62541.NewStringNodeID(c.Namespace, tagID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create the node id for Tag: %s", tagID))
		return err
	}

	v, err := open62541.NewVariant(value)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create the variant for Tag: %s", tagID))
		return err
	}
	c.iLog.Debug(fmt.Sprintf("Write Tag: %s Value: %v Success", tagID, v.Value()))
	return c.Client.WriteVariable(nodeID, v)
}

func (c *OPCClient) SubscribeGroup(tags []string, callback func(tagID string, value interface{})) error {
	c.iLog.Debug(fmt.Sprintf("Subscribe Group: %s", tags))
	subscription, err := c.Client.NewSubscription(100, 0)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create the subscription for Tags: %s", tags))
		return err
	}
	latestValues := make(map[string]interface{})
	for _, tagID := range tags {
		nodeID, err := open62541.NewStringNodeID(c.Namespace, tagID)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to create the node id for Tag: %s", tagID))
			return err
		}

		_, err = subscription.MonitorDataChange(nodeID, 0, func(v *open62541.Variant) {
			c.iLog.Debug(fmt.Sprintf("Tag '%s' value changed: %v\n", tagID, v.Value()))
			fmt.Printf("Tag '%s' value changed: %v\n", tagID, v.Value())
			latestValues[tagID] = v.Value()
			callback(tagID, v.Value())
		})
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to monitor the tag: %s", tagID))
			return err
		}
	}
	for tagID, value := range latestValues {
		callback(tagID, value)
	}
	c.iLog.Debug(fmt.Sprintf("Subscribe Group: %s created Success", tags))
	return nil
}
