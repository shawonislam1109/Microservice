package accounting

import (
	"context"
	"isp-management-system/internal/db"
	"isp-management-system/internal/models"
	"log"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// Job represents an accounting job to be processed by a worker.
type Job struct {
	Packet *radius.Packet
}

// Handler handles RADIUS accounting requests.
type Handler struct {
	jobQueue chan<- Job
}

// NewHandler creates a new accounting handler.
func NewHandler(jobQueue chan<- Job) *Handler {
	return &Handler{jobQueue: jobQueue}
}

// HandleAccountingRequest processes a RADIUS Accounting-Request.
func (h *Handler) HandleAccountingRequest(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	acctStatusType := rfc2866.AcctStatusType_Get(r.Packet)
	nasIP := rfc2865.NASIPAddress_Get(r.Packet)

	log.Printf("RADIUS: Received accounting request for user '%s' from NAS %s [Type: %s]", username, nasIP, acctStatusType)

	// Push the packet to the job queue for asynchronous processing.
	job := Job{Packet: r.Packet}
	select {
	case h.jobQueue <- job:
		// Job successfully queued
	default:
		log.Printf("RADIUS: Warning: Accounting job queue is full. Dropping packet for user '%s'", username)
	}

	// Immediately acknowledge the request to the MikroTik router.
	w.Write(r.Response(radius.CodeAccountingResponse))
}

// ProcessJob is executed by a worker to handle the database logic for a job.
func ProcessJob(job Job, repo db.Repository) {
	ctx := context.Background()
	p := job.Packet

	session := &models.Session{
		SessionID:        rfc2866.AcctSessionID_GetString(p),
		Username:         rfc2865.UserName_GetString(p),
		NASIPAddress:     rfc2865.NASIPAddress_Get(p).String(),
		CallingStationID: rfc2865.CallingStationID_GetString(p),
		InputOctets:      int64(rfc2866.AcctInputOctets_Get(p)),
		OutputOctets:     int64(rfc2866.AcctOutputOctets_Get(p)),
	}

	acctStatusType := rfc2866.AcctStatusType_Get(p)

	switch acctStatusType {
	case rfc2866.AcctStatusType_Value_Start:
		session.SessionStartTime = time.Now().UTC()
		if err := repo.CreateSession(ctx, session); err != nil {
			log.Printf("RADIUS Worker: Error creating session for user '%s': %v", session.Username, err)
		}

	case rfc2866.AcctStatusType_Value_Stop:
		session.SessionStopTime = time.Now().UTC()
		session.SessionTotalTime = int(rfc2866.AcctSessionTime_Get(p))
		session.TerminateCause = rfc2866.AcctTerminateCause_GetString(p)
		if err := repo.UpdateSession(ctx, session); err != nil {
			log.Printf("RADIUS Worker: Error updating session (stop) for user '%s': %v", session.Username, err)
		}

	case rfc2866.AcctStatusType_Value_Interim_Update:
		if err := repo.UpdateSession(ctx, session); err != nil {
			log.Printf("RADIUS Worker: Error updating session (interim) for user '%s': %v", session.Username, err)
		}
	}
}