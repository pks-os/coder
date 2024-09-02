package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/serpent"
)

func (r *RootCmd) logs() *serpent.Command {
	client := new(codersdk.Client)
	logsCmd := &serpent.Command{
		Annotations: workspaceCommand,
		Use:         "logs <workspace>",
		Short:       "Show logs for a given workspace build",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(1),
			r.InitClient(client),
		),
		Handler: logsHandler(client),
	}
	return logsCmd
}

func logsHandler(client *codersdk.Client) serpent.HandlerFunc {
	return func(inv *serpent.Invocation) error {
		ctx := inv.Context()
		if len(inv.Args) != 1 {
			return xerrors.New("must specify a single workspace!")
		}

		ws, err := namedWorkspace(inv.Context(), client, inv.Args[0])
		if err != nil {
			return err
		}

		agentIDs := make([]uuid.UUID, 0)
		for _, res := range ws.LatestBuild.Resources {
			for _, agt := range res.Agents {
				agentIDs = append(agentIDs, agt.ID)
			}
		}

		ch := make(chan string, 0)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range ch {
				_, _ = fmt.Fprintln(inv.Stdout, msg)
			}
		}()

		buildLogs, err := fetchBuildLogs(ctx, client, ws.LatestBuild.ID, time.Unix(0, 0))
		if err != nil {
			_, _ = fmt.Fprintf(inv.Stderr, "fetch build logs: %s", err.Error())
		}
		for bl := range buildLogs {
			ch <- bl
		}

		for _, agentID := range agentIDs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				agentLogs, err := fetchAgentLogs(ctx, client, agentID, time.Unix(0, 0), false)
				if err != nil {
					_, _ = fmt.Fprintf(inv.Stderr, "fetch agent logs: %s", err.Error())
				}
				for al := range agentLogs {
					ch <- al
				}
			}()
		}
		wg.Wait()
		return nil
	}
}

func fetchBuildLogs(ctx context.Context, client *codersdk.Client, buildID uuid.UUID, after time.Time) (chan string, error) {
	c := make(chan string)
	logCh, closeLogs, err := client.WorkspaceBuildLogsAfter(ctx, buildID, after.Unix())
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			_ = closeLogs.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				close(c)
				return
			case msg, ok := <-logCh:
				if !ok {
					return
				}
				c <- fmt.Sprintf("🏗️ %s [%s] (%s) %s", msg.CreatedAt, msg.Level, msg.Stage, msg.Output)
			}
		}
	}()
	return c, nil
}

func fetchAgentLogs(ctx context.Context, client *codersdk.Client, agentID uuid.UUID, after time.Time, follow bool) (chan string, error) {
	c := make(chan string)
	logCh, closeLogs, err := client.WorkspaceAgentLogsAfter(ctx, agentID, after.Unix(), follow)
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			_ = closeLogs.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				close(c)
				return
			case msgs, ok := <-logCh:
				if !ok {
					return
				}
				for _, msg := range msgs {
					c <- fmt.Sprintf("️🕵️ %s [%s] %s", msg.CreatedAt, msg.Level, msg.Output)
				}
			}
		}
	}()
	return c, nil
}
