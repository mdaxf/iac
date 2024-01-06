package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type WorkFlow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `json:"name"`
	UUID        string             `json:"uuid"`
	Version     string             `json:"version"`
	Description string             `json:"description"`
	ISDefault   bool               `json:"isDefault"`
	Type        string             `json:"type"`
	Nodes       []Node             `json:"nodes"`
	Links       []Link             `json:"links"`
}

type Node struct {
	Name          string                 `json:"name"`
	ID            string                 `json:"id"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	Page          string                 `json:"page"`
	TranCode      string                 `json:"trancode"`
	Roles         []string               `json:"roles"`
	Users         []string               `json:"users"`
	Roleids       []int64                `json:"roleids"`
	Userids       []int64                `json:"userids"`
	PreCondition  map[string]interface{} `json:"precondition"`
	PostCondition map[string]interface{} `json:"postcondition"`
	ProcessData   map[string]interface{} `json:"processdata"`
	RoutingTables []RoutingTable         `json:"routingtables"`
}

type Link struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Label  string `json:"label"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type RoutingTable struct {
	Default  bool   `json:"default"`
	Sequence int    `json:"sequence"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	Target   string `json:"target"`
}
