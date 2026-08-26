// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package blockcontroller

import (
	"strings"
	"testing"

	"github.com/wavetermdev/waveterm/pkg/waveobj"
)

// Requirement: auto-associate remote sessions with tmux.
// Acceptance criteria coverage:
// 1. Remote block with term:tmux:session → tmux attach is injected
// 2. Local block / block without the meta key → no injection
// 3. Session names are quote-escaped to prevent injection

func TestBuildTmuxAttachScript_RemoteConn(t *testing.T) {
	meta := waveobj.MetaMapType{
		waveobj.MetaKey_TermTmuxSession: "omlx-11335",
	}
	script := buildTmuxAttachScript(meta, "aws:co-gpu")
	if script == "" {
		t.Fatal("expected tmux attach script for remote conn with tmux session meta")
	}
	if !strings.Contains(script, `exec tmux new -A -t 'omlx-11335'`) {
		t.Fatalf("unexpected attach command: %q", script)
	}
	if !strings.Contains(script, `[ -n "$TMUX" ]`) {
		t.Fatalf("missing nested tmux guard: %q", script)
	}
	if !strings.Contains(script, `command -v tmux`) {
		t.Fatalf("missing tmux presence guard: %q", script)
	}
}

func TestBuildTmuxAttachScript_LocalConn(t *testing.T) {
	meta := waveobj.MetaMapType{
		waveobj.MetaKey_TermTmuxSession: "omlx-11335",
	}
	for _, connName := range []string{"local", "local:whatever", ""} {
		if script := buildTmuxAttachScript(meta, connName); script != "" {
			t.Fatalf("expected no script for local conn %q, got %q", connName, script)
		}
	}
}

func TestBuildTmuxAttachScript_NoMetaKey(t *testing.T) {
	if script := buildTmuxAttachScript(waveobj.MetaMapType{}, "aws:co-gpu"); script != "" {
		t.Fatalf("expected no script when meta key absent, got %q", script)
	}
	if script := buildTmuxAttachScript(waveobj.MetaMapType{waveobj.MetaKey_TermTmuxSession: ""}, "aws:co-gpu"); script != "" {
		t.Fatalf("expected no script when session name empty, got %q", script)
	}
}

func TestBuildTmuxAttachScript_EscapesSessionName(t *testing.T) {
	meta := waveobj.MetaMapType{
		waveobj.MetaKey_TermTmuxSession: `evil"; rm -rf /; echo "`,
	}
	script := buildTmuxAttachScript(meta, "aws:co-gpu")
	// Session names must be single-quoted so inner double quotes / semicolons cannot break shell structure.
	if !strings.Contains(script, `-t 'evil"; rm -rf /; echo "'`) {
		t.Fatalf("expected single-quoted session name, got: %q", script)
	}
	// Session names containing single quotes are also safe: inner single quotes escape to '\''.
	meta2 := waveobj.MetaMapType{
		waveobj.MetaKey_TermTmuxSession: `a'b`,
	}
	script2 := buildTmuxAttachScript(meta2, "aws:co-gpu")
	if !strings.Contains(script2, `-t 'a'\''b'`) {
		t.Fatalf("expected escaped inner single quote, got: %q", script2)
	}
}
