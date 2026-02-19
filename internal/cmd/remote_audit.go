package cmd

import "govard/internal/engine/remote"

func writeRemoteAuditEvent(event remote.AuditEvent) {
	_ = remote.WriteAuditEvent(event)
}
