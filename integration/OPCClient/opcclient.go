package opcclient

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
	"github.com/mdaxf/iac/logger"
)

type OPCConfig struct {
	OPCClients []OPCClient `json:"opcclients"`
}

type OPCClient struct {
	Client    *opcua.Client
	Endpoint  string        `json:"endpoint"`
	Host      string        `json:"host"`
	Name      string        `json:"name"`
	Namespace uint16        `json:"namespace"`
	CertFile  string        `json:"certFile"`
	KeyFile   string        `json:"keyFile"`
	Timeout   time.Duration `json:"timeout"`
	Nodes     map[string]*opcua.Node
	SubGroups []SubGroup `json:"subgroups"`
	iLog      logger.Log
}

type SubGroup struct {
	TriggerTags []string                    `json:"triggerTags"`
	ReportTags  []string                    `json:"reportTags"`
	Trigger     func(string, *ua.DataValue) `json:"trigger"`
	Report      func(string, *ua.DataValue) `json:"report"`
}

func Initialize(configurations OPCClient) {
	// Initialize the OPC UA client

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OPCClient"}

	iLog.Debug(fmt.Sprintf(("Create OPCClient with configuration : %s"), logger.ConvertJson(configurations)))

	opcclient := &OPCClient{
		Endpoint:  configurations.Endpoint,
		Namespace: configurations.Namespace,
		CertFile:  configurations.CertFile,
		KeyFile:   configurations.KeyFile,
		Timeout:   configurations.Timeout * time.Second,
		SubGroups: configurations.SubGroups,
		iLog:      iLog,
	}

	cancel := opcclient.CreateClient()
	opcclient.Connect()

	defer cancel()
	defer opcclient.Disconnect()
	defer opcclient.Client.CloseWithContext(context.Background())

	go func() {

		for _, subgroup := range opcclient.SubGroups {

			callbackfunc := func(tag string, v *ua.DataValue) {
				fmt.Printf("Tag: %v, Value: %v\n", tag, v.Value)
			}
			opcclient.Subscribe(subgroup.TriggerTags, callbackfunc)

		}
	}()
	// Wait for termination signal to gracefully shut down the client

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM)
	<-terminate
}

func (c *OPCClient) CreateClient() context.CancelFunc {
	// OPC UA server configuration
	var (
		endpoint = flag.String("endpoint", c.Endpoint, "OPC UA Endpoint URL")
		policy   = flag.String("policy", "", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		mode     = flag.String("mode", "", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
		certFile = flag.String("cert", "", "Path to cert.pem. Required for security mode/policy != None")
		keyFile  = flag.String("key", "", "Path to private key.pem. Required for security mode/policy != None")
	)
	ctx, cancel := context.WithCancel(context.Background())

	endpoints, err := opcua.GetEndpoints(ctx, *endpoint)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("error: %v", err))
	}

	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		c.iLog.Critical("Failed to find suitable endpoint")
	}
	c.iLog.Debug(fmt.Sprintf("Using endpoint: %v", ep.EndpointURL))
	c.iLog.Debug(fmt.Sprintf("%s,%s", ep.SecurityPolicyURI, ep.SecurityMode))
	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		opcua.CertificateFile(*certFile),
		opcua.PrivateKeyFile(*keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	opcclient := opcua.NewClient(ep.EndpointURL, opts...)

	c.Client = opcclient

	return cancel
}

func (c *OPCClient) BrowseEndpoints() (*ua.GetEndpointsResponse, error) {
	// Browse the endpoint by the server
	c.iLog.Debug(fmt.Sprintf("Browsing endpoints from server"))
	endpoints, err := c.Client.GetEndpoints()
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to get endpoints: %s", err))
		return nil, err
	}
	c.iLog.Debug(fmt.Sprintf("Endpoints: %v", endpoints))

	return endpoints, nil
}
func (c *OPCClient) Connect() {

	c.iLog.Debug(fmt.Sprintf("Connectting to OPC Server: %s", c.Endpoint))
	//	defer client.Close()

	// Connect to the OPC UA server
	err := c.Client.Connect(context.Background())
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to connect to OPC UA server: %s", err))
	}
	c.iLog.Debug(fmt.Sprintf("Connected to OPC Server: %s", c.Endpoint))
}

func (c *OPCClient) browseTags(nodeID string) ([]Tag, error) {
	// Browse the node
	c.iLog.Debug(fmt.Sprintf("Browsing nodes (tags) from server"))
	result, err := c.Client.Browse(&ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			&ua.BrowseDescription{
				NodeID:          ua.MustParseNodeID(nodeID),
				ReferenceTypeID: ua.NewNumericNodeID(0, 0),
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	})
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to browse nodes: %s", err))
		return nil, err
	}

	c.iLog.Debug(fmt.Sprintf("Browse result: %v", result))

	var tags []Tag

	// Process the browse result
	for _, res := range result.Results {
		for _, ref := range res.References {
			// Extract the tag name and address from the reference
			tagName := ref.DisplayName.Text
			tagAddress := ref.NodeID.String()

			// Create a new tag object
			tag := Tag{
				Name:    tagName,
				Address: tagAddress,
			}

			// Add the tag to the list
			tags = append(tags, tag)
		}
	}
	c.iLog.Debug(fmt.Sprintf("Tags: %v", tags))
	return tags, nil
}

func (c *OPCClient) readTagValue(nodeID string) (interface{}, error) {
	// Read the node value
	c.iLog.Debug(fmt.Sprintf("Reading node value: %s", nodeID))
	opcnodeid := flag.String("node", "", nodeID)

	id, err := ua.ParseNodeID(*opcnodeid)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("invalid node id: %v", err))
	}

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	var resp *ua.ReadResponse

	for {
		resp, err = c.Client.ReadWithContext(context.Background(), req)
		if err == nil {
			break
		}

		// Following switch contains known errors that can be retried by the user.
		// Best practice is to do it on read operations.
		switch {
		case err == io.EOF && c.Client.State() != opcua.Closed:
			// has to be retried unless user closed the connection
			time.After(1 * time.Second)
			continue

		case errors.Is(err, ua.StatusBadSessionIDInvalid):
			// Session is not activated has to be retried. Session will be recreated internally.
			time.After(1 * time.Second)
			continue

		case errors.Is(err, ua.StatusBadSessionNotActivated):
			// Session is invalid has to be retried. Session will be recreated internally.
			time.After(1 * time.Second)
			continue

		case errors.Is(err, ua.StatusBadSecureChannelIDInvalid):
			// secure channel will be recreated internally.
			time.After(1 * time.Second)
			continue

		default:
			c.iLog.Critical(fmt.Sprintf("Read failed: %s", err))
		}
	}

	if resp != nil && resp.Results[0].Status != ua.StatusOK {
		c.iLog.Critical(fmt.Sprintf("Status not OK: %v", resp.Results[0].Status))
	}

	return resp.Results[0].Value.Value(), nil
}

func (c *OPCClient) writeTagValue(nodeID string, value string) (ua.StatusCode, error) {

	opcnodeid := flag.String("node", "", nodeID)
	opcvalue := flag.String("value", "", value)

	// Create a variant from the value
	c.iLog.Debug(fmt.Sprintf("Writing node value: %s, %v", nodeID, value))
	id, err := ua.ParseNodeID(*opcnodeid)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("invalid node id: %v", err))
	}

	v, err := ua.NewVariant(*opcvalue)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("invalid value: %v", err))
	}

	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create variant: %s", err))
		return ua.StatusBad, err
	}

	// Write the node value
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := c.Client.WriteWithContext(context.Background(), req)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Write failed: %s", err))
		return ua.StatusBad, err
	}

	c.iLog.Debug(fmt.Sprintf("Node value written with result: %v", resp.Results[0]))
	return resp.Results[0], nil
}

func (c *OPCClient) Subscribe(triggerTags []string, callback func(string, *ua.DataValue)) {
	notifyCh := make(chan *opcua.PublishNotificationData, 1000)
	ctx := context.Background()
	interval := flag.Duration("interval", opcua.DefaultSubscriptionInterval, "subscription interval")
	sub, err := c.Client.SubscribeWithContext(ctx, &opcua.SubscriptionParameters{
		Interval: *interval,
	}, notifyCh)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to create subscription: %s", err))
	}
	defer sub.Cancel(ctx)

	event := flag.Bool("event", false, "subscribe to node event changes (Default: node value changes)")

	for _, tag := range triggerTags {
		// Create a monitored item for the tag
		nodeID := flag.String("node", tag, "node id to subscribe to")
		//nodeID := ua.NewStringNodeID(2, tag)
		id, err := ua.ParseNodeID(*nodeID)

		if err != nil {
			c.iLog.Critical(fmt.Sprintf("invalid node id: %v", err))
		}

		var miCreateRequest *ua.MonitoredItemCreateRequest
		var eventFieldNames []string

		if *event {
			miCreateRequest, eventFieldNames = eventRequest(id)
		} else {
			miCreateRequest = valueRequest(id)
		}
		res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
		if err != nil || res.Results[0].StatusCode != ua.StatusOK {
			c.iLog.Error(fmt.Sprintf("Failed to create monitored item for tag: %v", err))
		}
		// read from subscription's notification channel until ctx is cancelled
		for {
			select {
			case <-ctx.Done():
				return
			case res := <-notifyCh:
				if res.Error != nil {
					c.iLog.Error(fmt.Sprintf("there is error: %v", res.Error))
					continue
				}

				switch x := res.Value.(type) {
				case *ua.DataChangeNotification:
					for _, item := range x.MonitoredItems {
						data := item.Value.Value.Value()
						c.iLog.Debug(fmt.Sprintf("MonitoredItem with client handle %v = %v", item.ClientHandle, data))
					}

				case *ua.EventNotificationList:
					for _, item := range x.Events {
						c.iLog.Debug(fmt.Sprintf("Event for client handle: %v\n", item.ClientHandle))
						for i, field := range item.EventFields {
							c.iLog.Debug(fmt.Sprintf("%v: %v of Type: %T", eventFieldNames[i], field.Value(), field.Value()))
						}

					}

				default:
					c.iLog.Debug(fmt.Sprintf("what's this publish result? %T", res.Value))
				}
			}
		}

	}
}
func valueRequest(nodeID *ua.NodeID) *ua.MonitoredItemCreateRequest {
	handle := uint32(42)
	return opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
}

func eventRequest(nodeID *ua.NodeID) (*ua.MonitoredItemCreateRequest, []string) {
	fieldNames := []string{"EventId", "EventType", "Severity", "Time", "Message"}
	selects := make([]*ua.SimpleAttributeOperand, len(fieldNames))

	for i, name := range fieldNames {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}

	wheres := &ua.ContentFilter{
		Elements: []*ua.ContentFilterElement{
			{
				FilterOperator: ua.FilterOperatorGreaterThanOrEqual,
				FilterOperands: []*ua.ExtensionObject{
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
						},
						Value: ua.SimpleAttributeOperand{
							TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
							BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: "Severity"}},
							AttributeID:      ua.AttributeIDValue,
						},
					},
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.LiteralOperand_Encoding_DefaultBinary),
						},
						Value: ua.LiteralOperand{
							Value: ua.MustVariant(uint16(0)),
						},
					},
				},
			},
		},
	}

	filter := ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   wheres,
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	handle := uint32(42)
	req := &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:       nodeID,
			AttributeID:  ua.AttributeIDEventNotifier,
			DataEncoding: &ua.QualifiedName{},
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     handle,
			DiscardOldest:    true,
			Filter:           &filterExtObj,
			QueueSize:        10,
			SamplingInterval: 1.0,
		},
	}

	return req, fieldNames
}

/*
func (c *OPCClient) createSubscriptionGroup(triggerTags []string, callback func(string, *ua.DataValue)) (*opcua.Subscription, error) {

		sub, err := client.Subscribe(1*time.Second, ua.CreateSubscriptionRequest{
			RequestedPublishingInterval: 1000,
			RequestedMaxKeepAliveCount:  10,
			RequestedLifetimeCount:      100,
			PublishingEnabled:           true,
		})
		if err != nil {
			c.iLog.Critical(fmt.Sprintf("Failed to create subscription: %s", err))
		}

		for _,tag := range triggerTags{
			// Create a monitored item for the tag
			nodeID := ua.NewStringNodeID(2, tag)

			item, err := sub.MonitorValue(nodeID, ua.AttributeIDValue, func(v *ua.DataValue) {
				// Trigger action for tag1
				callback(tag, v)
				c.iLog.Debug(fmt.Sprintf("Trigger action for tag1: %v", v.Value))
			})
			if err != nil {
				c.iLog.Error(fmt.Sprintf("Failed to create monitored item for tag1: %v", err))
			}

		}

		return sub, nil
	}

	func (c *OPCClient)monitorSubscribeGroup(group []*opcua.Subscription) {
		filter  := flag.String("filter", "timestamp", "DataFilter: status, value, timestamp.")
		// Start the subscription
		for _, sub := range group {

			 need to change

			triggerNodeID := flag.String("trigger", "", "node id to trigger with")
			reportNodeID  := flag.String("report", "", "node id value to report on trigger")
			triggeringNode, err := ua.ParseNodeID(*triggerNodeID)
			if err != nil {
				c.iLog.Critical(fmt.Sprintf("There are error:%v",err))
			}

			triggeredNode, err := ua.ParseNodeID(*reportNodeID)
			if err != nil {
				c.iLog.Critical(fmt.Sprintf("There are error:%v",err))
			}

			miCreateRequests := []*ua.MonitoredItemCreateRequest{
				opcua.NewMonitoredItemCreateRequestWithDefaults(triggeringNode, ua.AttributeIDValue, 42),
				{
					ItemToMonitor: &ua.ReadValueID{
						NodeID:       triggeredNode,
						AttributeID:  ua.AttributeIDValue,
						DataEncoding: &ua.QualifiedName{},
					},
					MonitoringMode: ua.MonitoringModeSampling,
					RequestedParameters: &ua.MonitoringParameters{
						ClientHandle:     43,
						DiscardOldest:    true,
						Filter:           c.getFilter(*filter),
						QueueSize:        10,
						SamplingInterval: 0.0,
					},
				},
			}

			sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequests...)
		}
	}

func (c *OPCClient) getFilter(filterStr string) *ua.ExtensionObject {

		var filter ua.DataChangeFilter
		switch filterStr {
		case "status":
			filter = ua.DataChangeFilter{Trigger: ua.DataChangeTriggerStatus}
		case "value":
			filter = ua.DataChangeFilter{Trigger: ua.DataChangeTriggerStatusValue}
		case "timestamp":
			filter = ua.DataChangeFilter{Trigger: ua.DataChangeTriggerStatusValueTimestamp}
		default:
			log.Fatalf("Unable to match to a valid filter type: %v\nShould be status, value, or timestamp", filterStr)
		}

		return &ua.ExtensionObject{
			EncodingMask: ua.ExtensionObjectBinary,
			TypeID: &ua.ExpandedNodeID{
				NodeID: ua.NewNumericNodeID(0, id.DataChangeFilter_Encoding_DefaultBinary),
			},
			Value: filter,
		}
	}
*/
func (c *OPCClient) Disconnect() {
	// Disconnect from the OPC UA server
	c.Client.Close()
	c.iLog.Debug(fmt.Sprintf("Disconnected from OPC Server: %s", c.Endpoint))
}

type Tag struct {
	Name    string
	Address string
}
