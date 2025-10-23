package cdcmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"wikicrawler/internal/core/cdcmanager/cdc"
	"wikicrawler/internal/core/kafkaclient"
	"wikicrawler/internal/model"

	"github.com/jackc/pglogrepl"
)

type CDCManager struct {
	cdcclient     *cdc.CDCClient
	producers     map[string]*kafkaclient.KafkaProducer
	relationStore map[uint32]*pglogrepl.RelationMessage
}

func NewCDCManager(cfgcd *cdc.CDCConfig,
	cfgrelpro kafkaclient.KafkaConfig,
	cfgentpro kafkaclient.KafkaConfig,
	relTBname string,
	entTBname string) *CDCManager {

	cdcclient := cdc.NewCDCClient(cfgcd)
	relp := &kafkaclient.KafkaProducer{}
	relp.Init(cfgrelpro)
	entp := &kafkaclient.KafkaProducer{}
	entp.Init(cfgentpro)
	pros := make(map[string]*kafkaclient.KafkaProducer)
	pros[relTBname] = relp
	pros[entTBname] = entp

	return &CDCManager{
		cdcclient:     cdcclient,
		producers:     pros,
		relationStore: make(map[uint32]*pglogrepl.RelationMessage),
	}
}

func (c *CDCManager) Open() error {
	c.cdcclient.RegisterHandler(c.msgHandler)
	for name, producer := range c.producers {
		if err := producer.Open(); err != nil {
			return fmt.Errorf("[CDCManager] failed to open producer %s: %w", name, err)
		}
	}

	if err := c.cdcclient.Open(); err != nil {
		return err
	}
	return nil
}

func (c *CDCManager) Close() error {
	if err := c.cdcclient.Close(); err != nil {
		return err
	}
	for name, producer := range c.producers {
		if err := producer.Close(); err != nil {
			return fmt.Errorf("[CDCManager] failed to close producer %s: %w", name, err)
		}
	}
	return nil
}

func (c *CDCManager) msgHandler(logicalMsg pglogrepl.Message) {
	switch m := logicalMsg.(type) {

	case *pglogrepl.RelationMessage:
		c.relationStore[m.RelationID] = m
		fmt.Printf("🧩 Relation cached: %s.%s (cols: %d)\n",
			m.Namespace, m.RelationName, len(m.Columns))

	case *pglogrepl.InsertMessage:
		if rel, ok := c.relationStore[m.RelationID]; ok {
			c.onNewDBEvent("INSERT", rel, m.Tuple)
		} else {
			log.Printf("[CDCManager] unknown relation ID: %d", m.RelationID)
		}

	case *pglogrepl.UpdateMessage:
		if rel, ok := c.relationStore[m.RelationID]; ok {
			c.onNewDBEvent("UPDATE", rel, m.NewTuple)
		} else {
			log.Printf("[CDCManager] unknown relation ID: %d", m.RelationID)
		}

	case *pglogrepl.DeleteMessage:
		if rel, ok := c.relationStore[m.RelationID]; ok {
			c.onNewDBEvent("DELETE", rel, m.OldTuple)
		} else {
			log.Printf("[CDCManager] unknown relation ID: %d", m.RelationID)
		}

	case *pglogrepl.TruncateMessage:
		fmt.Printf("[CDCManager] TRUNCATE: %v\n", m.RelationIDs)

	default:
		fmt.Printf("[CDCManager] Unhandled logical message: %T\n", m)
	}
}

func (c *CDCManager) onNewDBEvent(op string, rel *pglogrepl.RelationMessage, tuple *pglogrepl.TupleData) {
	if rel == nil || tuple == nil {
		return
	}

	// Build a map of column name -> value
	rowData := make(map[string]string)
	for i, col := range tuple.Columns {
		if col.Data != nil {
			rowData[rel.Columns[i].Name] = string(col.Data)
		} else {
			rowData[rel.Columns[i].Name] = "NULL"
		}
	}

	// Create event struct
	dbevent := model.DBEvent{
		DBName: rel.RelationName,
		DBID:   rel.RelationID,
		CMD:    op,
		Data:   rowData,
	}

	// Log to console for debug
	fmt.Printf("CDC Event: %+v\n", dbevent)

	// Marshal to JSON
	data, err := json.Marshal(dbevent)
	if err != nil {
		fmt.Printf("Error marshalling DBEvent: %v\n", err)
		return
	}

	// Send event (e.g., to Kafka)
	if producer, ok := c.producers[rel.RelationName]; ok && producer != nil {
		producer.Push(data)
	} else {
		fmt.Printf("[CDCManager] No producer found for relation: %s\n", rel.RelationName)
	}
}
