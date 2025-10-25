package cdc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"wikicrawler/internal/utils/processor"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
)

type MessageHandler func(logicalMsg pglogrepl.Message)

type CDCConfig struct {
	Replicator       string
	Psw              string
	Address          string
	DB               string
	Replication_slot string
	OutputPlugin     string
	Publication      string
	Lsn              pglogrepl.LSN
}

//var relationStore = make(map[uint32]*pglogrepl.RelationMessage)

func (c *CDCConfig) ConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?replication=database",
		c.Replicator,
		c.Psw,
		c.Address,
		c.DB,
	)
}

type CDCClient struct {
	processor.BaseProcessor
	Config       *CDCConfig
	PostGresConn *pgconn.PgConn
	lastPingTime time.Time
	handler      MessageHandler
}

func NewCDCClient(Cfg *CDCConfig) *CDCClient {
	s := &CDCClient{
		Config: Cfg,
	}
	s.Init(s)
	return s
}

func (s *CDCClient) Open() error {
	fmt.Printf("[CDCClient] Connecting to: %s\n", s.Config.ConnStr())

	c, err := pgconn.Connect(context.Background(), s.Config.ConnStr())
	if err != nil {

		return err
	}
	s.PostGresConn = c

	// Create slot if needed (skip if already done)
	_, err = pglogrepl.CreateReplicationSlot(context.Background(),
		s.PostGresConn, s.Config.Replication_slot, s.Config.OutputPlugin, pglogrepl.CreateReplicationSlotOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	pluginArgs := []string{"proto_version '1'", fmt.Sprintf("publication_names '%s'", s.Config.Publication)}
	err = pglogrepl.StartReplication(context.Background(), s.PostGresConn, s.Config.Replication_slot, s.Config.Lsn, pglogrepl.StartReplicationOptions{PluginArgs: pluginArgs})
	if err != nil {
		return err
	}

	fmt.Printf("[CDCClient] connected!\n")

	return s.Start()
}

func (s *CDCClient) Close() error {
	// First, stop replication
	if err := s.Stop(); err != nil {
		return err
	}

	// Then close PostgreSQL connection
	if s.PostGresConn != nil {
		return s.PostGresConn.Close(context.Background())
	}

	return nil
}

func (s *CDCClient) RegisterHandler(h MessageHandler) {
	s.handler = h
}
func (s *CDCClient) RunningTask() {
	// Keep track of last time we sent a standby status update
	if s.lastPingTime.IsZero() {
		s.lastPingTime = time.Now()
	}

	// --- 1. Send heartbeat every 10s to confirm replication progress ---
	if time.Since(s.lastPingTime) > 10*time.Second {
		err := pglogrepl.SendStandbyStatusUpdate(
			context.Background(),
			s.PostGresConn,
			pglogrepl.StandbyStatusUpdate{
				WALWritePosition: s.Config.Lsn, // tell Postgres we have processed up to this LSN
				WALFlushPosition: s.Config.Lsn,
				WALApplyPosition: s.Config.Lsn,
			},
		)
		if err != nil {
			log.Printf("[CDCClient] failed to send standby status update: %v", err)
			return
		}
		log.Printf("[[CDCClient] - Heartbeat] confirmed up to LSN %s", s.Config.Lsn.String())
		s.lastPingTime = time.Now()
	}

	// --- 2. Wait for new replication message from server (10s timeout) ---
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg, err := s.PostGresConn.ReceiveMessage(ctxWithTimeout)
	if pgconn.Timeout(err) {
		// nothing received, loop again
		return
	}
	if err != nil {
		log.Printf("[CDCClient] error receiving message: %v", err)
		return
	}

	// --- 3. Parse replication message ---
	switch msg := msg.(type) {
	case *pgproto3.CopyData:
		switch msg.Data[0] {
		//sent by the PostgreSQL server
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			// It’s PostgreSQL telling your client:
			// “Hey, I’m alive, and here’s where my WAL stream currently is — do you still want me to continue sending data?”
			serverMsg, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				log.Printf("[CDCClient] failed to parse keepalive message: %v", err)
				return
			}
			if serverMsg.ReplyRequested {
				log.Println("[Keepalive] reply requested, sending standby status update...")
				err = pglogrepl.SendStandbyStatusUpdate(context.Background(), s.PostGresConn, pglogrepl.StandbyStatusUpdate{
					WALWritePosition: s.Config.Lsn,
					WALFlushPosition: s.Config.Lsn,
					WALApplyPosition: s.Config.Lsn,
				})
				if err != nil {
					log.Printf("[CDCClient] failed to reply keepalive: %v", err)
				}
			}

		case pglogrepl.XLogDataByteID:
			// --- 4. Parse logical replication WAL data ---
			xlog, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				log.Printf("[CDCClient] failed to parse xlog data: %v", err)
				return
			}

			// Decode logical message (INSERT / UPDATE / DELETE)
			logicalMsg, err := pglogrepl.Parse(xlog.WALData)
			if err != nil {
				log.Printf("[CDCClient] failed to parse logical msg: %v", err)
				return
			}

			s.handler(logicalMsg)
			// --- 5. Update LSN after processing message ---
			s.Config.Lsn = xlog.WALStart + pglogrepl.LSN(len(xlog.WALData))
			//log.Printf("[Advance LSN] now at %s", s.Config.Lsn.String())
		}
	default:
		log.Printf("[CDCClient] unexpected message: %T", msg)
	}
}
